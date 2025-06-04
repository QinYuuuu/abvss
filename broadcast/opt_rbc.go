package broadcast

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/QinYuuuu/abvss/crypto/erasurecode"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"google.golang.org/protobuf/proto"

	"github.com/QinYuuuu/abvss/crypto/hasher"
)

/*
   Implementation of Validated Reliable Broadcast from DXL21 with good case optimization.
   Briefly, the protocol proceeds as follows:
   1. Broadcaster sends the proposal to all
   2. Nodes run Bracha's RBC on hash
   3. OptRBC i output once the RBC on hash terminates and if it has received a matching proposal from leader
   4. Otherwise, node i triggers a fallback protocol that uses ADD to help node i recover the proposal.
*/

const (
	Propose int64 = iota
	Echo
	Ready
	Terminate
	ADDTrigger
	ADDDisperse
	ADDReconstruct
)

type OptRBC struct {
	pid     int64
	n, f    int64
	running bool
	// session state
	sessions map[string]*Session    // sessionID -> session state
	output   map[string]chan []byte // sessionID -> output channel
	// network interface
	send    func(int64, *protobuf.OptRBCMessage)
	receive func() (*protobuf.OptRBCMessage, bool)
}

func NewOptRBC(pid, n, f int64, send func(int64, *protobuf.OptRBCMessage), receive func() (*protobuf.OptRBCMessage, bool)) *OptRBC {
	return &OptRBC{
		pid:      pid,
		n:        n,
		f:        f,
		running:  false,
		sessions: make(map[string]*Session),
		send:     send,
		output:   make(map[string]chan []byte),
		receive:  receive,
	}
}

func (n *OptRBC) CreateNewSession(sessionID string, leader int64) {
	n.output[sessionID] = make(chan []byte, 1)
	n.sessions[sessionID] = NewSession(n.pid, leader, n.n, n.f, sessionID, n.send)
	n.sessions[sessionID].output = n.output[sessionID]
}

func (n *OptRBC) StartNewBroadcast(msg []byte, leader int64, sessionID string) {
	if n.pid == leader {
		n.sessions[sessionID].initiateBroadcast(msg)
	}
}

func (n *OptRBC) Output(sessionID string) chan []byte {
	return n.output[sessionID]
}

type Session struct {
	pid       int64
	sessionID string
	leader    int64
	n, f      int64

	// protocol state
	stripes               [][]byte
	echoCounter           map[string]*int64
	readyCounter          map[string]*int64
	addDisperseCounter    map[string]*int64
	echoSenders           map[int64]bool
	readySenders          map[int64]bool
	terminateSenders      map[int64]bool
	addTriggerSenders     map[int64]bool
	addDisperseSenders    map[int64]bool
	addReconstructSenders map[int64]bool

	// message store
	leaderHash        []byte
	reconstructedHash []byte
	leaderMsg         []byte
	reconstructedMsg  []byte
	committedHash     []byte

	// signal
	readySent    bool
	addReadySent bool
	committed    bool
	outputted    bool

	output chan []byte
	encode func([]byte) [][]byte
	// network interface
	send func(int64, *protobuf.OptRBCMessage)
}

func NewSession(pid, leader, nNodes, f int64, sessionID string, send func(int64, *protobuf.OptRBCMessage)) *Session {
	encoder := erasurecode.NewReedSolomonCode(int(nNodes-2*f), int(nNodes))
	encode := func(input []byte) [][]byte {
		chunks, err := encoder.Encode(input)
		if err != nil {
			return nil
		}
		result := make([][]byte, len(chunks))
		for i, chunk := range chunks {
			result[i] = chunk.GetData()
		}
		return result
	}
	return &Session{
		pid:                   pid,
		sessionID:             sessionID,
		leader:                leader,
		n:                     nNodes,
		f:                     f,
		send:                  send,
		encode:                encode,
		echoCounter:           make(map[string]*int64),
		readyCounter:          make(map[string]*int64),
		addDisperseCounter:    make(map[string]*int64),
		echoSenders:           make(map[int64]bool),
		readySenders:          make(map[int64]bool),
		terminateSenders:      make(map[int64]bool),
		addTriggerSenders:     make(map[int64]bool),
		addDisperseSenders:    make(map[int64]bool),
		addReconstructSenders: make(map[int64]bool),
		committed:             false,
		readySent:             false,
		addReadySent:          false,
		stripes:               nil,
		reconstructedMsg:      nil,
		committedHash:         nil,
		reconstructedHash:     nil,
		leaderHash:            nil,
		leaderMsg:             nil,
	}
}

func (s *Session) Output() chan []byte {
	return s.output
}

func (s *Session) initiateBroadcast(msg []byte) {
	slog.Debug(fmt.Sprintf("[node %v] [RBC: %v] start broadcast on %v", s.pid, s.sessionID, msg))
	if s.leader != s.pid {
		slog.Info("only leader send propose")
		return
	}
	s.leaderMsg = msg

	for i := range s.n {
		proposeMsg := &protobuf.OptRBCMessage{
			FromID:    s.pid,
			DestID:    i,
			SessionID: s.sessionID,
			MsgType:   Propose,
			Payload:   msg,
		}
		//slog.Info(fmt.Sprintf("[node %v] [RBC: %v] send message %v", s.pid, s.sessionID, msg))
		s.send(i, proposeMsg)
	}
}

func (n *OptRBC) Run() {
	if n.running {
		return
	} else {
		go n.messageLoop()
		n.running = true
	}
}

func (n *OptRBC) messageLoop() {
	for {
		if msg, ok := n.receive(); ok {
			slog.Debug(fmt.Sprintf("[node %v] [RBC: %v] receive %v message from %v", n.pid, msg.SessionID, msg.MsgType, msg.FromID))
			s := n.sessions[msg.SessionID]
			if s == nil {
				slog.Error(fmt.Sprintf("[node %v] [RBC: %v] not exist", n.pid, msg.SessionID))
			}
			switch msg.MsgType {
			case Propose:
				// handle propose
				s.handlePropose(msg.FromID, msg.Payload)
			case Echo:
				// handle echo
				s.handleEcho(msg.FromID, msg.Payload)
			case Ready:
				// handle ready
				s.handleReady(msg.FromID, msg.Payload)
			case ADDTrigger:
				// handle addTrigger
				s.handleADDTrigger(msg.FromID, msg.Payload)
			case ADDDisperse:
				// handle addDisperse
				slog.Info(fmt.Sprintf("[node %v] [RBC: %v] addDisperse: %v", n.pid, n.sessions, msg.Payload))
				s.handleADDDisperse(msg.FromID, msg.Payload)
			case ADDReconstruct:
				// handle addReconstruct
				slog.Info(fmt.Sprintf("[node %v] [RBC: %v] addRecomstruct: %v", n.pid, n.sessions, msg.Payload))
				s.handleADDReconstruct(msg.FromID, msg.Payload)
			default:
				slog.Error("unhandled", slog.Any("type", msg.MsgType))
				panic("unhandled default case")
			}
		}
	}
}

func (s *Session) handlePropose(sender int64, msg []byte) {
	if sender != s.leader {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] receive message from node %v not leader", s.pid, s.sessionID, sender))
		return
	}
	if s.committed {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] have received propose message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] [RBC: %v] handle Propose message from %v", s.pid, s.sessionID, sender))
	digest := hasher.MD5Hasher(msg)
	s.leaderMsg = msg
	s.leaderHash = digest
	for i := range s.n {
		echoMsg := &protobuf.OptRBCMessage{
			FromID:    s.pid,
			DestID:    i,
			SessionID: s.sessionID,
			MsgType:   Echo,
			Payload:   digest,
		}
		s.send(i, echoMsg)
	}
}

func (s *Session) handleEcho(sender int64, payload []byte) {
	if s.echoSenders[sender] {
		slog.Info("leader is", slog.Any("id", s.leader))
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] has received ECHO message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] [RBC: %v] handle Echo message from %v", s.pid, s.sessionID, sender))
	digest := payload

	s.echoSenders[sender] = true
	if s.echoCounter[string(digest)] == nil {
		s.echoCounter[string(digest)] = new(int64)
		atomic.StoreInt64(s.echoCounter[string(digest)], 0)
	}
	atomic.AddInt64(s.echoCounter[string(digest)], 1)

	if atomic.LoadInt64(s.echoCounter[string(digest)]) >= 2*s.f+1 && !s.readySent {
		s.readySent = true
		for i := range s.n {
			readyMsg := &protobuf.OptRBCMessage{
				FromID:    s.pid,
				DestID:    i,
				SessionID: s.sessionID,
				MsgType:   Ready,
				Payload:   digest,
			}
			s.send(i, readyMsg)
		}
	}
}

func (s *Session) handleReady(sender int64, payload []byte) {
	// check the sender
	if s.readySenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] has received Ready message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] [RBC: %v] handle Ready message from %v", s.pid, s.sessionID, sender))
	digest := payload
	s.readySenders[sender] = true

	if s.readyCounter[string(digest)] == nil {
		s.readyCounter[string(digest)] = new(int64)
		atomic.StoreInt64(s.readyCounter[string(digest)], 0)
	}
	atomic.AddInt64(s.readyCounter[string(digest)], 1)
	if atomic.LoadInt64(s.readyCounter[string(digest)]) >= s.f+1 {
		s.committedHash = digest
		if bytes.Equal(digest, s.leaderHash) {
			s.committed = true
			if !s.outputted {
				s.output <- s.leaderMsg
				s.outputted = true
				//slog.Info(fmt.Sprintf("[node %v] [RBC: %v] output message %v", s.pid, s.sessionID, s.leaderMsg))
			}
		} else {
			for i := range s.n {
				if i == s.pid {
					continue
				}
				addTriggerMsg := &protobuf.OptRBCMessage{
					FromID:    s.pid,
					DestID:    i,
					SessionID: s.sessionID,
					MsgType:   ADDTrigger,
					Payload:   nil,
				}
				s.send(i, addTriggerMsg)
			}
		}
	}
}

func (s *Session) handleADDTrigger(sender int64, payload []byte) {
	if s.addTriggerSenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] has received ADDTrigger message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] [RBC: %v] handle ADDTrigger message from %v", s.pid, s.sessionID, sender))
	s.addTriggerSenders[sender] = true
	if s.committed {
		if s.stripes == nil {
			s.stripes = s.encode(s.leaderMsg)
		}
		disperseData := &protobuf.DisperseData{
			MyStripe:     s.stripes[s.pid],
			SenderStripe: s.stripes[sender],
		}
		disperseDataBytes, err := proto.Marshal(disperseData)
		if err != nil {
			slog.Error("proto marshal", slog.Any("error", err))
			return
		}
		addDisperseMsg := &protobuf.OptRBCMessage{
			FromID:    s.pid,
			DestID:    sender,
			SessionID: s.sessionID,
			MsgType:   ADDDisperse,
			Payload:   disperseDataBytes,
		}
		s.send(sender, addDisperseMsg)
	}
}

func (s *Session) handleADDDisperse(sender int64, payload []byte) {
	if s.addTriggerSenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] has received ADDDisperse message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] [RBC: %v] handle ADDDisperse message from %v", s.pid, s.sessionID, sender))
	s.addDisperseSenders[sender] = true

	var disperseData protobuf.DisperseData
	err := proto.Unmarshal(payload, &disperseData)
	if err != nil {
		slog.Error("proto marshal", slog.Any("error", err))
		return
	}
	if s.addDisperseCounter[string(disperseData.SenderStripe)] == nil {
		s.echoCounter[string(disperseData.SenderStripe)] = new(int64)
		atomic.StoreInt64(s.echoCounter[string(disperseData.SenderStripe)], 0)
	}
	atomic.AddInt64(s.echoCounter[string(disperseData.SenderStripe)], 1)
	s.stripes[s.pid] = disperseData.MyStripe
	if atomic.LoadInt64(s.readyCounter[string(disperseData.SenderStripe)]) >= s.f+1 && !s.addReadySent {
		s.addReadySent = true
		for i := range s.n {
			addReadyMsg := &protobuf.OptRBCMessage{
				FromID:    s.pid,
				DestID:    i,
				SessionID: s.sessionID,
				MsgType:   ADDReconstruct,
				Payload:   disperseData.SenderStripe,
			}
			s.send(i, addReadyMsg)
		}
	}
}

func (s *Session) handleADDReconstruct(sender int64, payload []byte) {
	if s.addReconstructSenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] [RBC: %v] has received ADDReconstruct message from node %v", s.pid, s.sessionID, sender))
		return
	}
}
