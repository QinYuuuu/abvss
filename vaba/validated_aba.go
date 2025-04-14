package vaba

import (
	"context"
	"fmt"
	"log"
)

const (
	ACSS   string = "A"
	RBC    string = "R"
	ABA    string = "B"
	PREKEY string = "P"
	KEY    string = "K"
	RA     string = "T"
)

// VABAImpl 实现验证的异步拜占庭协议
type VABAImpl struct {
	raInput       chan interface{}
	indexInput    [][]int
	validSet      map[int]bool
	rbc0Signal    chan struct{}
	rbc0Signals   []chan struct{}
	proposedValue [][]byte
	g2            interface{} // 这里需要根据实际类型替换

	terminateView int

	output            int
	tSignal           chan struct{}
	acssOutputs       map[int]map[string]interface{}
	acssSignal        chan struct{}
	keyProposal       [][]int
	keyProposalSignal []chan struct{}

	expView    int
	publicKeys interface{} // 这里需要根据实际类型替换
	privateKey interface{} // 这里需要根据实际类型替换
	g          interface{} // 这里需要根据实际类型替换
	h          interface{} // 这里需要根据实际类型替换
	n          int
	t          int
	deg        int
	myID       int
	sc0        int
	sc         int
	send       func(int, interface{})
	recv       chan interface{}
	pc         interface{} // 这里需要根据实际类型替换
	ZR         interface{} // 这里需要根据实际类型替换
	G1         interface{} // 这里需要根据实际类型替换
	multiexp   interface{} // 这里需要根据实际类型替换
	dotprod    interface{} // 这里需要根据实际类型替换
	poly       interface{} // 这里需要根据实际类型替换

	subscribeRecv func(string) chan interface{}
	matrix        interface{} // 这里需要根据实际类型替换

	getSend     func(string) func(int, interface{})
	outputQueue chan interface{}
	logger      *log.Logger
	mks         map[int]bool
	//acss          *ASKS
	acssTask     context.CancelFunc
	acssTaskList []context.CancelFunc
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewVABA 创建新的VABA实例
func NewVABA(publicKeys interface{}, privateKey interface{}, g interface{}, g2 interface{}, h interface{},
	n, t, deg, myID int, send func(int, interface{}), recv chan interface{}, pc interface{},
	curveParams []interface{}, matrices interface{}) *VABAImpl {

	ctx, cancel := context.WithCancel(context.Background())

	v := &VABAImpl{
		raInput:           make(chan interface{}, n),
		indexInput:        make([][]int, n),
		validSet:          make(map[int]bool),
		rbc0Signal:        make(chan struct{}),
		rbc0Signals:       make([]chan struct{}, n),
		proposedValue:     make([][]byte, n),
		g2:                g2,
		publicKeys:        publicKeys,
		privateKey:        privateKey,
		g:                 g,
		h:                 h,
		n:                 n,
		t:                 t,
		deg:               deg,
		myID:              myID,
		sc0:               0,
		expView:           1,
		keyProposal:       make([][]int, n),
		keyProposalSignal: make([]chan struct{}, n),
		send:              send,
		recv:              recv,
		pc:                pc,
		ZR:                curveParams[0],
		G1:                curveParams[1],
		multiexp:          curveParams[2],
		dotprod:           curveParams[3],
		matrix:            matrices,
		outputQueue:       make(chan interface{}, 1),
		logger:            log.New(log.Writer(), fmt.Sprintf("[Node %d] ", myID), log.LstdFlags),
		ctx:               ctx,
		cancel:            cancel,
	}

	v.sc = v.sc0 + v.expView

	// 初始化通道
	for i := 0; i < n; i++ {
		v.rbc0Signals[i] = make(chan struct{})
		v.keyProposalSignal[i] = make(chan struct{})
	}

	return v
}
