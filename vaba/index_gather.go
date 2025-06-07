package vaba

import (
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

type IGImpl struct {
	id, n, t   int64
	instanceID string
	valid      sync.Map // Set of validated parties
	validSize  int64
	ci         map[int64]bool // Set of indices for valid PREPARE messages

	informSent      bool
	prepareReceived map[int64][]int64
	ackReceived     map[int64]bool // Track which parties sent ACKs
	ackCounter      int64
	output          chan []int64 // Final output X_i
	terminated      bool

	// Communication channels
	send    func(message *protobuf.IGMessage)
	receive func() chan *protobuf.IGMessage
}

func NewIGImpl(id, n, t int64, instanceID string, send func(message *protobuf.IGMessage), receive func() chan *protobuf.IGMessage) *IGImpl {
	return &IGImpl{
		id:              id,
		n:               n,
		t:               t,
		instanceID:      instanceID,
		ci:              make(map[int64]bool),
		ackReceived:     make(map[int64]bool),
		prepareReceived: make(map[int64][]int64),
		output:          make(chan []int64, 10),
		terminated:      false,
		send:            send,
		receive:         receive,
	}
}

func (p *IGImpl) GetValid() []int64 {
	var Si []int64
	p.valid.Range(func(key any, _ any) bool {
		Si = append(Si, key.(int64))
		return true
	})
	return Si
}

// AddValid adds a party to the Valid set
func (p *IGImpl) AddValid(index int64) {
	_, hasStore := p.valid.LoadOrStore(index, true)
	if !hasStore {
		atomic.AddInt64(&p.validSize, 1)
	}
	slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v]", p.id, p.instanceID), slog.Any("add valid", index), slog.Any("valid size", atomic.LoadInt64(&p.validSize)), slog.Any("inform send", p.informSent))
	if !p.informSent && atomic.LoadInt64(&p.validSize) == p.n-p.t {
		p.informSent = true
		Si := p.GetValid()
		// Send INFORM message to all parties
		var i int64
		for i = 0; i < p.n; i++ {
			msg := &protobuf.IGMessage{
				FromID:     p.id,
				DestID:     i,
				InstanceID: p.instanceID,
				Type:       "INFORM",
				Set:        Si,
			}
			if i == p.id {
				p.receive() <- msg
				continue
			}
			p.send(msg)
		}
		p.informSent = true
	}
}

func (p *IGImpl) Input(valid []int64) {
	for _, i := range valid {
		p.AddValid(i)
	}
}

func (p *IGImpl) Run() {
	// Process incoming messages
	go func() {
		for {
			select {
			case msg := <-p.receive():
				p.ProcessMessage(msg)
			}
		}
	}()
}

func (p *IGImpl) ProcessMessage(msg *protobuf.IGMessage) {
	if p.terminated {
		return
	}
	switch msg.Type {
	case "INFORM":
		// Line 4-5: upon S_j ⊆ Valid_i becomes true, send <ACK> to party j
		// loop through S_j and check if each element is in Valid_i
		Sj := msg.Set
		isSubset := true
		for _, j := range Sj {
			if _, contain := p.valid.Load(j); !contain {
				validList := p.GetValid()
				slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v] receive %v, not in valid %v", p.id, p.instanceID, j, validList))
				isSubset = false
				p.receive() <- msg
				break
			}
		}
		if isSubset {
			// Send ACK to the sender
			newMsg := &protobuf.IGMessage{
				FromID:     p.id,
				DestID:     msg.FromID,
				InstanceID: p.instanceID,
				Type:       "ACK",
			}
			p.send(newMsg)
		}

	case "ACK":
		// Record that we received an ACK from this party
		sender := msg.FromID
		if p.ackReceived[sender] {
			return
		}
		p.ackReceived[sender] = true
		p.ackCounter++
		// Line 6-8: upon receiving ACK from n-t distinct nodes
		if p.ackCounter == p.n-p.t {
			slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v] get %v ACK", p.id, p.instanceID, p.ackCounter))
			// Let T_i := Valid_i
			Ti := p.GetValid()
			// Send PREPARE to all
			var i int64
			for i = 0; i < p.n; i++ {
				newMsg := &protobuf.IGMessage{
					FromID:     p.id,
					DestID:     i,
					InstanceID: p.instanceID,
					Type:       "PREPARE",
					Set:        Ti,
				}
				p.send(newMsg)
			}
		}

	case "PREPARE":
		slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v] handle Prepare from %v", p.id, p.instanceID, msg.FromID), slog.Any("T_i", msg.Set))
		// Line 10-13: upon T_j ⊆ Valid_i becomes true
		Tj := msg.Set
		p.prepareReceived[msg.FromID] = Tj
		isSubset := true
		for _, j := range Tj {
			if _, contain := p.valid.Load(j); !contain {
				validList := p.GetValid()
				slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v] receive %v, not in valid %v", p.id, p.instanceID, j, validList))
				isSubset = false
				p.receive() <- msg
				break
			}
		}
		if isSubset {
			// Line 11: C_i := C_i ∪ {j}
			p.ci[msg.FromID] = true
			slog.Info(fmt.Sprintf("[node %v] [IndexGather: %v] get subset", p.id, p.instanceID), slog.Any("ci", p.ci))
			// Line 12-13: if |C_i| = n-t then output X_i := ⋃_{j∈C_i} T_j and terminate
			if int64(len(p.ci)) == p.n-p.t {
				// Compute union of all T_j
				unionMap := make(map[int64]bool)
				for j := range p.ci {
					// We need to find the T_j set for each j in C_i
					// For simplicity, we'll use what we received
					T_j := p.prepareReceived[j]
					for _, item := range T_j {
						unionMap[item] = true
					}
				}
				result := make([]int64, 0, len(unionMap))
				for item := range unionMap {
					result = append(result, item)
				}
				slog.Info(fmt.Sprintf("[node %d] [IndexGather: %v] IGImpl terminated with output: %v\n", p.id, p.instanceID, result))
				p.output <- result
				p.terminated = true
			}
		}
	}
}

func (p *IGImpl) Output() chan []int64 {
	return p.output
}
