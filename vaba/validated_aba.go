package vaba

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"sync"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	ACSS   string = "A"
	RBC    string = "R"
	ABA    string = "B"
	PREKEY string = "P"
	KEY    string = "K"
	RA     string = "T"
)

type asksOutput struct {
	index int64
}

// VABAImpl implement validated aba
type VABAImpl struct {
	raInput       chan interface{}
	rbc0Signal    chan struct{}
	rbc0Signals   []chan struct{}
	proposedValue [][]byte

	terminateView int

	output            int
	tSignal           chan struct{}
	acssOutputs       map[int]map[string]interface{}
	acssSignal        chan struct{}
	keyProposal       [][]int
	keyProposalSignal []chan struct{}

	n, t, myID    int64
	validSet      sync.Map
	asksCounter   int64
	asksSharedSet []int64
	proposedSet   []int64

	send func(int, interface{})
	recv chan interface{}

	subscribeRecv func(string) chan interface{}

	getSend     func(string) func(int, interface{})
	outputQueue chan interface{}
	mks         map[int]bool

	asksInstances []*ASKSImpl
	icgInstance   *IndexCoverGatherImpl
	rbc           *broadcast.OptRBC
	//acss          *ASKS
	acssTask     context.CancelFunc
	acssTaskList []context.CancelFunc
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewVABA 创建新的VABA实例
func NewVABA(nodeNum, threshold, myID int64, send func(int, interface{}), recv chan interface{}, pc interface{},
	curveParams []interface{}, matrices interface{}) *VABAImpl {

	ctx, cancel := context.WithCancel(context.Background())

	v := &VABAImpl{
		raInput:           make(chan interface{}, nodeNum),
		rbc0Signal:        make(chan struct{}),
		rbc0Signals:       make([]chan struct{}, nodeNum),
		proposedValue:     make([][]byte, nodeNum),
		n:                 nodeNum,
		t:                 threshold,
		myID:              myID,
		asksCounter:       0,
		keyProposal:       make([][]int, nodeNum),
		keyProposalSignal: make([]chan struct{}, nodeNum),
		send:              send,
		recv:              recv,
		outputQueue:       make(chan interface{}, 1),
		ctx:               ctx,
		cancel:            cancel,
	}

	return v
}

func (vaba *VABAImpl) Run() {
	for _, asks := range vaba.asksInstances {
		asks.Run()
	}
	vaba.asksInstances[vaba.myID].Share()
	vaba.icgInstance.Run()
	vaba.rbc.Run()
	vaba.rbc.CreateNewSession("pre_"+strconv.FormatInt(vaba.myID, 10), vaba.myID)
	go vaba.messageLoop()
}

func (vaba *VABAImpl) Valid(index int64) {
	_, hasStore := vaba.validSet.LoadOrStore(index, true)
	if !hasStore {
		slog.Info(fmt.Sprintf("[node %v][vaba] Valid %v", vaba.myID, index))
	}
}

func (vaba *VABAImpl) messageLoop() {
	rbcOutputChan := vaba.getRBCPreOutput()
	for {
		select {
		case msg := <-vaba.recv:
			slog.Info(fmt.Sprintf("[node %v] receive", vaba.myID), slog.Any("msg", msg))
			// case output
		case rbcOutput := <-rbcOutputChan:
			slog.Info(fmt.Sprintf("[node %v] output in RBC_%v: %v", vaba.myID, rbcOutput.index, rbcOutput.message))
			vaba.handleRBCPreOutput(rbcOutput, rbcOutputChan)
		case asksOutput := <-vaba.getASKSOutput():
			slog.Info(fmt.Sprintf("[node %v] output in ASKS %v: %v", vaba.myID, asksOutput.index, asksOutput))
			vaba.handleASKSOutput(asksOutput)
		}
	}

}

func (vaba *VABAImpl) handleASKSOutput(output *asksOutput) {
	vaba.asksCounter++
	vaba.asksSharedSet = append(vaba.asksSharedSet, output.index)
	// |Shared_i|= t + 1 for the first time
	if vaba.asksCounter == vaba.t+1 {
		sharedSet := make([]int64, vaba.t+1)
		copy(sharedSet, vaba.asksSharedSet)
		sharedSetBytes := vaba.convertSharedSetToBytes(sharedSet)
		sessionID := "pre_" + strconv.FormatInt(vaba.myID, 10)
		vaba.rbc.StartNewBroadcast(sharedSetBytes, vaba.myID, sessionID)
	}

}

func (vaba *VABAImpl) handleRBCPreOutput(output *rbcOutput, outputChan chan *rbcOutput) {
	pjBytes := output.message
	pj := vaba.convertBytesToSharedSet(pjBytes)
	slog.Info(fmt.Sprintf("[node %v] output in RBC_%v: %v", vaba.myID, output.index, pj))
	// check p_j ⊆ Valid_i
	isSubset := true
	for _, index := range pj {
		_, hasStore := vaba.validSet.Load(index)
		if !hasStore {
			isSubset = false
			outputChan <- output
			break
		}
	}
	if isSubset {
		// valid index in index cover gather
		vaba.icgInstance.ValidateParty(output.index)
	}
}

type rbcOutput struct {
	index   int64
	message []byte
}

func (vaba *VABAImpl) convertSharedSetToBytes(asksSharedSet []int64) []byte {
	proposeSetMsg := &protobuf.HartsProposeSet{
		Index: asksSharedSet,
	}
	proposeSetBytes, err := proto.Marshal(proposeSetMsg)
	if err != nil {
		slog.Error("Marshal harts propose set failed", slog.String("err", err.Error()))
	}
	return proposeSetBytes
}

func (vaba *VABAImpl) convertBytesToSharedSet(proposeSetBytes []byte) []int64 {
	var proposeSet protobuf.HartsProposeSet
	err := proto.Unmarshal(proposeSetBytes, &proposeSet)
	if err != nil {
		slog.Error("Unmarshal harts propose set failed", slog.String("err", err.Error()))
	}
	return proposeSet.Index
}

func (vaba *VABAImpl) getASKSOutput() chan *asksOutput {
	finishChan := make(chan *asksOutput, vaba.n-vaba.t)
	for i := int64(0); i < vaba.n; i++ {
		go func(sessionId int64) {
			output := <-vaba.asksInstances[sessionId].Output()
			if output != nil {
				slog.Info(fmt.Sprintf("[node %v] output in ASKS %v: %v", vaba.myID, sessionId, output))
				finishChan <- &asksOutput{
					index: sessionId,
				}
			}
		}(i)
	}
	return finishChan
}

func (vaba *VABAImpl) getRBCPreOutput() chan *rbcOutput {
	finishChan := make(chan *rbcOutput, vaba.n)
	for i := int64(0); i < vaba.n; i++ {
		go func(index int64) {
			sessionID := "pre_" + strconv.FormatInt(index, 10)
			output := <-vaba.rbc.Output(sessionID)
			if output != nil {
				slog.Info(fmt.Sprintf("[node %v] output in RBC_%v: %v", vaba.myID, sessionID, output))
				finishChan <- &rbcOutput{
					index:   index,
					message: output,
				}
			}
		}(i)
	}
	return finishChan
}
