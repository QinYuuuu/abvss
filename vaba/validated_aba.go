package vaba

import (
	"context"
	"fmt"
	"log/slog"
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
	indexInput    [][]int
	validSet      map[int]bool
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

	expView int

	nodeNum, threshold, myID int64
	asksCounter              int64
	asksSharedSet            []int64
	proposedSet              []int64
	sc0                      int
	sc                       int
	send                     func(int, interface{})
	recv                     chan interface{}

	subscribeRecv func(string) chan interface{}

	getSend     func(string) func(int, interface{})
	outputQueue chan interface{}
	mks         map[int]bool

	asksInstances []*ASKSImpl
	icgInstance   *IndexCoverGatherImpl
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
		indexInput:        make([][]int, nodeNum),
		validSet:          make(map[int]bool),
		rbc0Signal:        make(chan struct{}),
		rbc0Signals:       make([]chan struct{}, nodeNum),
		proposedValue:     make([][]byte, nodeNum),
		nodeNum:           nodeNum,
		threshold:         threshold,
		myID:              myID,
		sc0:               0,
		expView:           1,
		keyProposal:       make([][]int, nodeNum),
		keyProposalSignal: make([]chan struct{}, nodeNum),
		send:              send,
		recv:              recv,
		outputQueue:       make(chan interface{}, 1),
		ctx:               ctx,
		cancel:            cancel,
	}

	v.sc = v.sc0 + v.expView

	return v
}

func (vaba *VABAImpl) Run() {
	for _, asks := range vaba.asksInstances {
		asks.Run()
	}
	vaba.icgInstance.Run()

	asksFinishChan := vaba.getASKSOutput()
	for {
		asksFinish := <-asksFinishChan
		if asksFinish != nil {
			vaba.asksCounter++
			vaba.asksSharedSet = append(vaba.asksSharedSet, asksFinish.index)
			if vaba.asksCounter == vaba.threshold+1 {

			}
		}
	}
}

func (vaba *VABAImpl) getASKSOutput() chan *asksOutput {
	finishChan := make(chan *asksOutput, vaba.nodeNum-vaba.threshold)
	var i int64
	for i = 0; i < vaba.nodeNum; i++ {
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
