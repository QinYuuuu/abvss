package vaba

import (
	"fmt"
	"log/slog"

	"github.com/QinYuuuu/abvss/pkg/protobuf"
)

type RAImpl struct {
	n, id, t   int64
	instanceID string
	myInput    []byte
	output     chan []byte

	terminated   bool
	echoSender   map[int64]bool
	readySender  map[int64]bool
	echoCounter  map[string]int64 // Content -> SenderIDs
	readyCounter map[string]int64 // Content -> SenderIDs
	readySent    bool
	hasOutput    bool

	send    func(*protobuf.RAMessage)
	receive func() chan *protobuf.RAMessage
}

func NewRAImpl(id, n, t int64, instanceID string) *RAImpl {
	return &RAImpl{
		id:           id,
		n:            n,
		t:            t,
		instanceID:   instanceID,
		echoSender:   make(map[int64]bool),
		readySender:  make(map[int64]bool),
		echoCounter:  make(map[string]int64),
		readyCounter: make(map[string]int64),
		readySent:    false,
		output:       make(chan []byte),
	}
}

func (p *RAImpl) Output() chan []byte {
	return p.output
}

// Input party p_i input m_i
func (p *RAImpl) Input(input []byte) {
	p.myInput = input
	// Send <ECHO, m_i> to all
	p.SendToAll(ECHO, input)
}

// SendToAll Send a message to all parties
func (p *RAImpl) SendToAll(msgType string, content []byte) {
	var i int64
	for i = 0; i < p.n; i++ {
		msg := &protobuf.RAMessage{
			FromID:     p.id,
			DestID:     i,
			InstanceID: p.instanceID,
			Type:       msgType,
			Value:      content,
		}
		if i == p.n {
			p.receive() <- msg
		} else {
			p.send(msg)
		}
	}
}

// Run process received messages
func (p *RAImpl) Run() {
	go func() {
		for {
			select {
			case msg := <-p.receive():
				//slog.Info(fmt.Sprintf("[node %v] [session %v] ReliableAgreementImpl recv %v from %v", p.id, msg.InstanceID, msg.Type, msg.FromID))
				p.handleMessage(msg)
			}
		}
	}()
}

func (p *RAImpl) handleMessage(msg *protobuf.RAMessage) {
	switch msg.Type {
	case ECHO:
		p.handleEcho(msg)
	case READY:
		p.handleReady(msg)
	}
}

// Handle ECHO message
func (p *RAImpl) handleEcho(msg *protobuf.RAMessage) {
	content := msg.Value
	senderID := msg.FromID

	// Record that we received an ECHO from this sender for this content
	if p.echoSender[senderID] {
		return
	}
	p.echoSender[senderID] = true
	p.echoCounter[string(content)]++
	// Check if we've received ECHO from n-t parties for this content
	if p.echoCounter[string(content)] >= p.n-p.t && !p.readySent {
		// Send <READY, m> to all
		p.SendToAll(READY, content)
		p.readySent = true
	}
}

// Handle READY message
func (p *RAImpl) handleReady(msg *protobuf.RAMessage) {
	content := msg.Value
	senderID := msg.FromID

	// Record that we received a READY from this sender for this content
	if p.readySender[senderID] {
		return
	}
	p.readySender[senderID] = true
	p.readyCounter[string(content)]++
	// Check if we've received READY from t+1 parties for this content
	if p.readyCounter[string(content)] >= p.t+1 && !p.readySent {
		// Send <READY, m> to all
		p.SendToAll(READY, content)
		p.readySent = true
	}

	// Check if we've received READY from n-t parties for this content
	if p.readyCounter[string(content)] >= p.n-p.t && !p.hasOutput {
		// Output m and terminate
		p.output <- content
		p.hasOutput = true
		p.terminated = true
		slog.Info(fmt.Sprintf("[node %v] [RA: %s] outputs: %s", p.id, p.instanceID, content))
	}
}
