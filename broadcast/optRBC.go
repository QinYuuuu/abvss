package broadcast

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/QinYuuuu/abvss/crypto/hasher"
)

/*
   Implementation of Validated Reliable Broadcast from DXL21 with good case optimization.
   Briefly, the protocol proceeds as follows:
   1. Broadcaster sends the proposal to all
   2. Nodes run Bracha's RBC on hash
   3. Node i output once the RBC on hash terminates and if it has received a matching proposal from leader
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

type Message struct {
	FromID    int64
	DestID    int64
	SessionID int64
	MsgType   int64
	Payload   []byte
}

type Node struct {
	pid  int64
	n, f int64
	// session state
	sessions map[int64]*Session // sessionID -> session state
	output   []chan []byte      // sessionID -> output channel
	// network interface
	send    func(int64, Message)
	receive func() (Message, bool)
}

func NewNode(pid, n, f int64, send func(int64, Message), output []chan []byte, receive func() (Message, bool)) *Node {
	if output == nil {
		slog.Error("output is nil")
	}
	return &Node{
		pid:      pid,
		n:        n,
		f:        f,
		sessions: make(map[int64]*Session),
		send:     send,
		output:   output,
		receive:  receive,
	}
}

func (n *Node) CreateNewSession(sessionID, leader int64) {
	n.sessions[sessionID] = NewSession(n.pid, sessionID, leader, n.n, n.f, n.output[sessionID], n.send)
}

func (n *Node) StartNewBroadcast(msg []byte, leader, nNodes, f, sessionID int64) {
	if n.pid == leader {
		n.sessions[sessionID].initiateBroadcast(msg)
	}
}

func (n *Node) Output(sessionID int64) []byte {
	if data, ok := <-n.output[sessionID]; ok {
		return data
	} else {
		slog.Error("output channel closed")
		return nil
	}
}

type Session struct {
	pid       int64
	sessionID int64
	leader    int64
	n, f      int64

	// protocol state
	stripes      [][]byte
	echoCounter  map[string]*int64
	readyCounter map[string]*int64

	echoSenders           map[int64]bool
	readySenders          map[int64]bool
	terminateSenders      map[int]struct{}
	addTriggerSenders     map[int]struct{}
	addDisperseSenders    map[int]struct{}
	addReconstructSenders map[int]struct{}
	addDisperseCounter    map[string]int

	// message store
	leaderHash        []byte
	reconstructedHash []byte
	leaderMsg         []byte
	reconstructedMsg  []byte
	committedHash     []byte

	// 标志位
	readySent    bool
	addReadySent bool
	committed    bool

	// 外部接口
	output chan []byte

	// network interface
	send func(int64, Message)
}

func NewSession(pid, sessionID, leader, nNodes, f int64, output chan []byte, send func(int64, Message)) *Session {
	return &Session{
		pid:                   pid,
		sessionID:             sessionID,
		leader:                leader,
		n:                     nNodes,
		f:                     f,
		output:                output,
		send:                  send,
		echoCounter:           make(map[string]*int64),
		readyCounter:          make(map[string]*int64),
		echoSenders:           make(map[int64]bool),
		readySenders:          make(map[int64]bool),
		terminateSenders:      make(map[int]struct{}),
		addTriggerSenders:     make(map[int]struct{}),
		addDisperseSenders:    make(map[int]struct{}),
		addReconstructSenders: make(map[int]struct{}),
		addDisperseCounter:    make(map[string]int),
		committed:             false,
		readySent:             false,
		addReadySent:          false,
		stripes:               make([][]byte, 0),
		reconstructedMsg:      nil,
		committedHash:         nil,
		reconstructedHash:     nil,
		leaderHash:            nil,
		leaderMsg:             nil,
	}
}

func (s *Session) initiateBroadcast(msg []byte) {
	slog.Info(fmt.Sprintf("[node %v] session[%v] start broadcast", s.pid, s.sessionID))
	if s.leader != s.pid {
		slog.Info("only leader send propose")
		return
	}
	s.leaderMsg = msg

	for i := range s.n {
		proposeMsg := Message{
			FromID:    s.pid,
			DestID:    i,
			SessionID: s.sessionID,
			MsgType:   Propose,
			Payload:   msg,
		}
		//slog.Info(fmt.Sprintf("[node %v] session[%v] send message %v", s.pid, s.sessionID, msg))
		s.send(i, proposeMsg)
	}
}

func (n *Node) Run() {
	go n.messageLoop()
}

func (n *Node) messageLoop() {
	for {
		if msg, ok := n.receive(); ok {
			//slog.Info(fmt.Sprintf("[node %v] session[%v] receive %v message from %v", n.pid, msg.SessionID, msg.MsgType, msg.FromID))
			s := n.sessions[msg.SessionID]
			if s == nil {
				slog.Error(fmt.Sprintf("[node %v] session[%v] not exist", n.pid, msg.SessionID))
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
			default:
				panic("unhandled default case")
			}
		}
	}
}

func (s *Session) handlePropose(sender int64, msg []byte) {
	if sender != s.leader {
		slog.Info(fmt.Sprintf("[node %v] session[%v] receive message from node %v not leader", s.pid, s.sessionID, sender))
		return
	}
	if s.committed {
		slog.Info(fmt.Sprintf("[node %v] session[%v] have received propose message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] session[%v] handle Propose message from %v", s.pid, s.sessionID, sender))
	digest := hasher.MD5Hasher(msg)
	s.leaderMsg = msg
	for i := range s.n {
		echoMsg := Message{
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
		slog.Info(fmt.Sprintf("[node %v] session[%v] has received ECHO message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] session[%v] handle Echo message from %v", s.pid, s.sessionID, sender))
	digest := payload

	s.echoSenders[sender] = true
	if s.echoCounter[string(digest)] == nil {
		s.echoCounter[string(digest)] = new(int64)
		atomic.StoreInt64(s.echoCounter[string(digest)], 0)
	}
	atomic.AddInt64(s.echoCounter[string(digest)], 1)

	if atomic.LoadInt64(s.echoCounter[string(digest)]) >= s.f && !s.readySent {
		s.readySent = true
		for i := range s.n {
			readyMsg := Message{
				FromID:    s.pid,
				DestID:    i,
				SessionID: s.sessionID,
				MsgType:   Ready,
				Payload:   nil,
			}
			s.send(i, readyMsg)
		}
	}
}

func (s *Session) handleReady(sender int64, payload []byte) {
	// check the sender
	if s.readySenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] session[%v] has received Ready message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] session[%v] handle Ready message from %v", s.pid, s.sessionID, sender))
	digest := payload
	s.readySenders[sender] = true

	if s.readyCounter[string(digest)] == nil {
		s.readyCounter[string(digest)] = new(int64)
		atomic.StoreInt64(s.readyCounter[string(digest)], 0)
	}
	atomic.AddInt64(s.readyCounter[string(digest)], 1)

	if atomic.LoadInt64(s.readyCounter[string(digest)]) >= s.f {
		s.committedHash = digest
		if bytes.Equal(digest, s.leaderHash) {
			s.committed = true
			s.output <- s.leaderMsg
		} else {
			for i := range s.n {
				addTriggerMsg := Message{
					FromID:    s.pid,
					DestID:    i,
					SessionID: s.sessionID,
					MsgType:   ADDDisperse,
					Payload:   nil,
				}
				s.send(i, addTriggerMsg)
			}
		}
	}
}

func (s *Session) handleADDTrigger(sender int64, payload []byte) {
	if s.addTriggerSenders[sender] {
		slog.Info(fmt.Sprintf("[node %v] session[%v] has received ADDTrigger message from node %v", s.pid, s.sessionID, sender))
		return
	}
	slog.Info(fmt.Sprintf("[node %v] session[%v] handle ADDTrigger message from %v", s.pid, s.sessionID, sender))
	s.addTriggerSenders[sender] = true
	if len(s.addTriggerSenders) >= s.f {
		for i := range s.n {
			addDisperseMsg := Message{
				FromID:    s.pid,
				DestID:    i,
				SessionID: s.sessionID,
				MsgType:   ADDDisperse,
				Payload:   nil,
			}
			s.send(i, addDisperseMsg)
		}
	}
}