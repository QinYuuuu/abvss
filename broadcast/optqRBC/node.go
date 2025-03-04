package optqRBC

import (
	"abvss/crypto/erasurecode"
	"abvss/crypto/hasher"
	"bytes"
	"log"
	"log/slog"
	"sync"
	"sync/atomic"
)

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
	Payload   interface{}
}

type Node struct {
	pid int64
	// session state
	sessions     sync.Map // map[int64]*Session
	curSessionID atomic.Int64

	// network interface
	send    func(int64, any)
	receive func() (int, any)
}

func (n *Node) StartNewBroadcast(msg []byte, leader int64, nNodes, f int64) int64 {
	sessionID := n.curSessionID.Add(1)
	session := &Session{
		sessionID: sessionID,
		leader:    leader,
		n:         nNodes,
		f:         f,
	}

	n.sessions.Store(sessionID, session)

	if n.pid == leader {
		go session.initiateBroadcast(msg)
	}

	return sessionID
}

func (s *Session) handlePropose(sender int64, msg []byte) {
	if sender != s.leader || s.committed {
		log.Printf("[node %v] session[%v] receive message from node %v not leader", s.sessionID, sender)
		return
	}

	if !s.predicate(msg) {
		log.Printf("[node %v] session[%d] Invalid proposal", s.sessionID)
		return
	}

	digest := hasher.MD5Hasher(msg)
	echoMsg := Message{
		SessionID: s.sessionID,
		MsgType:   Echo,
		Payload:   digest,
	}

	for i := range s.n {
		s.send(i, echoMsg)
	}
}

func (s *Session) handleEcho(sender int64, payload []byte) {
	if s.echoSenders[sender] {
		log.Printf("[node %v][session %v] has received ECHO message from node %v", s.pid, s.sessionID, sender)
		return
	}
	digest := payload

	s.echoSenders[sender] = true
	s.echoCounter[string(digest)].Add(1)

	if s.echoCounter[string(digest)].Load() >= s.f && !s.readySent {
		s.readySent = true
		readyMsg := Message{
			SessionID: s.sessionID,
			MsgType:   Ready,
			Payload:   nil,
		}
		for i := range s.n {
			s.send(i, readyMsg)
		}
	}
}

func (s *Session) handleReady(sender int64, payload []byte) {
	// check the sender
	if s.readySenders[sender] {
		log.Printf("[node %v][session %v] has received ECHO message from node %v", s.pid, s.sessionID, sender)
		return
	}

	digest := payload
	s.readySenders[sender] = true
	s.readyCounter[string(digest)].Add(1)

	if s.readyCounter[string(digest)].Load() >= s.f {
		s.committedHash = digest
		if bytes.Equal(digest, s.leaderHash) {
			s.committed = true
		} /* else if bytes.Equal(digest, s.reconstructedHash) {
			s.commit(s.reconstructedMsg)
		} */else {

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

func (s *Session) handleAddTrigger(sender int64, payload []byte) {
	if s.committed {
		return
	}
	s.stripes = make([][]byte, s.n)
	if s.leader == s.pid {
		rscode := erasurecode.NewReedSolomonCode(int(s.f), int(s.pid))
		chunks, err := rscode.Encode(s.leaderMsg)
		if err != nil {
			slog.Error("")
		}
		for i := range s.n {
			disperseMsg := Message{
				FromID:    s.pid,
				DestID:    i,
				SessionID: s.sessionID,
				MsgType:   ADDDisperse,
				Payload:   chunks,
			}
			s.send(i, disperseMsg)
		}

	}
}
