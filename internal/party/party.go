package party

import (
	"errors"
	"log"
	"net"
	"sync"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/pkg/core"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
=======
	"abvss/pkg/core"
	"abvss/pkg/protobuf"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/pkg/core"
	"abvss/pkg/protobuf"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
)

// Party is a interface of consensus parties
type Party interface {
	send(m *protobuf.Message, des uint32) error
	broadcast(m *protobuf.Message) error
	getMessageWithType(messageType string) (*protobuf.Message, error)
}

// HonestParty is a struct of honest consensus parties
type HonestParty struct {
	N                 uint32
	F                 uint32
	PID               uint32
	ipList            []string
	portList          []string
	sendChannels      []chan *protobuf.Message
	dispatcheChannels *sync.Map
	lis               *net.TCPListener
	conns             []*net.TCPConn
	Bandwidth         uint64

	SigPK *share.PubPoly  //tss pk
	SigSK *share.PriShare //tss sk

	EncPK kyber.Point       //tse pk
	EncVK []*share.PubShare //tse vk
	EncSK *share.PriShare   //tse sk
}

// NewHonestParty return a new honest party object
func NewHonestParty(N uint32, F uint32, pid uint32, ipList, portList []string, sigPK *share.PubPoly, sigSK *share.PriShare, encPK kyber.Point, encVK []*share.PubShare, encSK *share.PriShare) *HonestParty {
	p := HonestParty{
		N:            N,
		F:            F,
		PID:          pid,
		ipList:       ipList,
		portList:     portList,
		sendChannels: make([]chan *protobuf.Message, N),
		conns:        make([]*net.TCPConn, N),
		Bandwidth:    0,

		SigPK: sigPK,
		SigSK: sigSK,

		EncPK: encPK,
		EncVK: encVK,
		EncSK: encSK,
	}

	return &p
}

func (p *HonestParty) Close() {
	for _, conn := range p.conns {
		err := conn.Close()
		if err != nil {
			log.Println("close honest party conn error:", err)
		}
	}
	err := p.lis.Close()
	if err != nil {
		log.Println("close honest party lis error:", err)
	}
}

// InitReceiveChannel setup the listener and Init the receiveChannel
func (p *HonestParty) InitReceiveChannel() error {
	lis, receivechan := core.MakeReceiveChannel(p.portList[p.PID])
	p.lis = lis
	p.dispatcheChannels = core.MakeDispatcheChannels(receivechan, p.N)
	return nil
}

// InitSendChannel setup the sender and Init the sendChannel, please run this after initializing all party's receiveChannel
func (p *HonestParty) InitSendChannel() error {
	for i := uint32(0); i < p.N; i++ {
		p.conns[i], p.sendChannels[i] = core.MakeSendChannel(p.ipList[i], p.portList[i])
	}
	return nil
}

// Send a message to party des
func (p *HonestParty) Send(m *protobuf.Message, des uint32) error {
	if !p.checkInit() {
		return errors.New("this party hasn't been initialized")
	}
	if des < p.N {
		p.sendChannels[des] <- m
		p.Bandwidth += uint64(len(m.GetData()))
		return nil
	}
	return errors.New("Destination id is too large")
}

// Broadcast a message to all parties
func (p *HonestParty) Broadcast(m *protobuf.Message) error {
	if !p.checkInit() {
		return errors.New("this party hasn't been initialized")
	}
	for i := uint32(0); i < p.N; i++ {
		err := p.Send(m, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMessage Try to get a message according to messageType, ID
func (p *HonestParty) GetMessage(messageType string, ID []byte) chan *protobuf.Message {
	value1, _ := p.dispatcheChannels.LoadOrStore(messageType, new(sync.Map))

	var value2 any
	if messageType == "Dec" {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, p.N*p.N))
	} else {
		value2, _ = value1.(*sync.Map).LoadOrStore(string(ID), make(chan *protobuf.Message, p.N))
	}

	return value2.(chan *protobuf.Message)
}

func (p *HonestParty) checkInit() bool {
	if p.sendChannels == nil {
		return false
	}
	return true
}
