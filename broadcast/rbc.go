package broadcast

import (
	"errors"
	"log"

	"github.com/QinYuuuu/abvss/crypto/hasher"
)

const SEND = "S"
const ECHO = "E"
const READY = "R"

var ErrUnknownMessageType = errors.New("unknown protocol message type")

type RBC struct {
	id, n, f   int64
	instanceID int64
	leader     bool
	*State
}

type State struct {
	Data            []byte
	Echos, Readys   int64
	nEchos, nReadys []bool
	SentReady       bool
	Output          bool
}

func NewRBC(id, n, f, instanceID int64) *RBC {
	return &RBC{id: id, n: n, f: f, instanceID: instanceID, State: NewState(n)}
}

func (r *RBC) SetLeader() {
	r.leader = true
}

func (r *RBC) Send(data []byte) ([]RBCMessage, error) {
	if !r.leader {
		return nil, errors.New("not leader cannot send")
	}
	msgs := make([]RBCMessage, r.n)
	var i int64
	for i = 0; i < r.n; i++ {
		msgs[i] = RBCMessage{FromID: r.id, DestID: i, MsgType: SEND, Data: data}
	}
	return msgs, nil
}

func (r *RBC) Recv(m RBCMessage) ([]RBCMessage, error) {
	if r.instanceID != m.InstanceID {
		return nil, errors.New("wrong InstanceID")
	}
	if m.DestID != r.id {
		return nil, errors.New("wrong receiver id")
	}
	log.Printf("node %v receive %v from node %v", r.id, m.MsgType, m.FromID)
	var msgs []RBCMessage
	switch m.MsgType {
	case SEND:
		if r.Data != nil {
			return nil, errors.New("duplicate send message")
		}
		r.Data = m.Data
		var i int64
		for i = 0; i < r.n; i++ {
			msg := RBCMessage{FromID: r.id, DestID: i, Data: hasher.SHA256Hasher(r.Data), MsgType: ECHO, InstanceID: r.instanceID}
			msgs = append(msgs, msg)
		}
	case ECHO:

		if r.Data == nil {
			return nil, errors.New("invalid echo message, no send")
		}
		if !r.nEchos[m.FromID] {
			r.nEchos[m.FromID] = true
			r.Echos++
		} else {
			return nil, nil
		}

	case READY:
		if r.Data == nil {
			return nil, errors.New("invalid ready message, no send")
		}
		if !r.nReadys[m.FromID] {
			r.nReadys[m.FromID] = true
			r.Readys++
		} else {
			return nil, nil
		}
	default:
		return nil, ErrUnknownMessageType
	}
	if r.Echos >= (r.n+r.f+1)/2 {
		var i int64
		for i = 0; i < r.n; i++ {
			msg := RBCMessage{
				FromID:     r.id,
				DestID:     i,
				Data:       hasher.SHA256Hasher(r.Data),
				MsgType:    READY,
				InstanceID: m.InstanceID,
			}
			msgs = append(msgs, msg)
		}
	}
	if r.Readys >= r.f+1 {
		var i int64
		for i = 0; i < r.n; i++ {
			msg := RBCMessage{FromID: r.id, DestID: i, Data: hasher.SHA256Hasher(r.Data), MsgType: READY, InstanceID: r.instanceID}
			msgs = append(msgs, msg)
		}
	}
	if r.Readys >= 2*r.f+1 {
		r.Output = true
	}
	log.Printf("node %v nECHOs %v nREADYs %v", r.id, r.Echos, r.Readys)
	return msgs, nil
}

// NewState creates a new protocol state based on an incoming message from a client
func NewState(n int64) *State {
	state := &State{
		Echos:     0,
		Readys:    0,
		nEchos:    make([]bool, n),
		nReadys:   make([]bool, n),
		SentReady: false,
		Output:    false,
	}
	return state
}
