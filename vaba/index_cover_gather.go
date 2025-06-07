package vaba

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

const (
	RaMsg      string = "RA_MSG"     // For reliable agreement messages
	WITHDRAW   string = "WITHDRAW"   // For withdraw notifications
	VALIDATION string = "VALIDATION" // For party validation notifications
)

// IndexCoverGatherImpl represents a participant in the protocol
type IndexCoverGatherImpl struct {
	id, n, t  int64 // Fault tolerance threshold
	sessionID string
	// State for Index Cover Gather Protocol
	valid            map[int64]bool // Set of parties that pi has validated so far
	withdraw         bool           // Whether party has withdrawn
	withdrawReceived map[int64]bool // Tracks received WITHDRAW messages
	// Reliable Agreement instances
	reliableAgreementInstances []*RAImpl
	indexGatherInstance        *IGImpl
	// Communication channels
	send    func(message *protobuf.ICGMessage)
	receive func() chan *protobuf.ICGMessage

	// Final output
	output        chan []int64 // Set Xi (indices of validated parties)
	withdrawReady chan bool
	terminated    bool // Whether protocol has terminated
}

// NewIndexCoverGatherImpl creates a new party instance for the protocol
func NewIndexCoverGatherImpl(id, n, t int64, sessionId string, send func(message *protobuf.ICGMessage), receive func() chan *protobuf.ICGMessage) *IndexCoverGatherImpl {
	reliableAgreementInstances := make([]*RAImpl, n)
	for j := int64(0); j < n; j++ {
		reliableAgreementInstances[j] = NewRAImpl(id, n, t, "ICG"+strconv.FormatInt(j, 10))
	}
	return &IndexCoverGatherImpl{
		id:                         id,
		n:                          n,
		t:                          t,
		sessionID:                  sessionId,
		valid:                      make(map[int64]bool),
		reliableAgreementInstances: reliableAgreementInstances,
		output:                     make(chan []int64, 1),
		withdraw:                   false,
		withdrawReceived:           make(map[int64]bool),
		withdrawReady:              make(chan bool),
		send:                       send,
		receive:                    receive,
		terminated:                 false,
	}
}

// ValidateParty is called when party_i has validated party_j
func (p *IndexCoverGatherImpl) ValidateParty(j int64) {
	// Skip if already validated
	if p.valid[j] {
		return
	}
	// Add to validated set
	p.valid[j] = true
	slog.Info(fmt.Sprintf("[node %v] [IndexCoverGather: %v] validated party %d", p.id, p.sessionID, j))

	// If not withdrawn yet, provide input 1 to corresponding RA instance
	if !p.withdraw {
		p.reliableAgreementInstances[j].Input([]byte("1"))
		// Simulate RA protocol by broadcasting the input
		slog.Info(fmt.Sprintf("[node %v] [IndexCoverGather: %v] input 1 to RA[%v]", p.id, p.sessionID, p.reliableAgreementInstances[j].instanceID))
	}
}

func (p *IndexCoverGatherImpl) getRAOutput() chan int64 {
	finishChan := make(chan int64, p.n)
	var i int64
	for i = 0; i < p.n; i++ {
		go func(sessionId int64) {
			output := <-p.reliableAgreementInstances[sessionId].Output()
			if output != nil {
				slog.Info(fmt.Sprintf("[node %v] output in RA %v", p.id, p.reliableAgreementInstances[sessionId].instanceID))
				finishChan <- sessionId
			}
		}(i)
	}
	return finishChan
}

// HandleWithdrawMessage processes a WITHDRAW message
func (p *IndexCoverGatherImpl) HandleWithdrawMessage(msg *protobuf.ICGMessage) {
	senderID := msg.FromID
	p.withdrawReceived[senderID] = true
	slog.Info(fmt.Sprintf("[node %d] [IndexCoverGather: %v] received WITHDRAW from party %d", p.id, p.sessionID, senderID), slog.Any("withdraw received", len(p.withdrawReceived)))
	// Check if we've received WITHDRAW from n-t parties
	if int64(len(p.withdrawReceived)) >= p.n-p.t && !p.terminated {
		p.terminated = true
		p.withdrawReady <- true
	}
}

// Run starts the party's protocol execution
func (p *IndexCoverGatherImpl) Run() {
	for _, ra := range p.reliableAgreementInstances {
		ra.Run()
	}
	p.indexGatherInstance.Run()
	slog.Info(fmt.Sprintf("[node %d] [IndexCoverGather: %v] start", p.id, p.sessionID))
	go func() {
		for {
			select {
			case msg := <-p.receive():
				switch msg.Type {
				case WITHDRAW:
					p.HandleWithdrawMessage(msg)
				}
			}
		}
	}()

	go func() {
		raFinishChan := p.getRAOutput()
		for {
			select {
			case index := <-raFinishChan:
				p.indexGatherInstance.AddValid(index)
				if int64(len(p.indexGatherInstance.GetValid())) == p.n-p.t {
					p.withdraw = true
					slog.Info(fmt.Sprintf("[node %v] [IndexCoverGather: %v] broadcast withdraw", p.id, p.sessionID))
					for i := int64(0); i < p.n; i++ {
						msg := &protobuf.ICGMessage{
							FromID:     p.id,
							DestID:     i,
							InstanceID: p.sessionID,
							Type:       WITHDRAW,
						}
						if msg.DestID == p.id {
							p.receive() <- msg
							continue
						}
						p.send(msg)
					}
				}
			}
		}
	}()
	go func() {
		for {
			select {
			case xi := <-p.indexGatherInstance.Output():
				slog.Info(fmt.Sprintf("[node %v] index gather output", p.id))
				p.output <- xi
			}
		}
	}()
}

func (p *IndexCoverGatherImpl) Output() chan []int64 {
	slog.Info(fmt.Sprintf("[node %v] index cover gather waiting withdraw ready", p.id))
	<-p.withdrawReady
	slog.Info(fmt.Sprintf("[node %v] index cover gather withdraw ready", p.id))
	return p.output
}
