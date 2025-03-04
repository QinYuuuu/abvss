package rbc

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"errors"
	"github.com/QinYuuuu/abvss/crypto/hasher"
=======
	"abvss/crypto/hasher"
	"errors"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/crypto/hasher"
	"errors"
>>>>>>> 19b0d27 (Initial commit)
	"log"
)

const SEND = "S"
const ECHO = "E"
const READY = "R"

var ErrUnknownMessageType = errors.New("unknown protocol message type")

type RBC struct {
	id, n, f   int
	instanceID int
	leader     bool
	*State
}

type State struct {
	Data            []byte
	Echos, Readys   int
	nEchos, nReadys []bool
	SentReady       bool
	Output          bool
}
type Message struct {
	instanceID            int
	fromID, destID, index int
	mtype                 string
	data                  []byte
}

func NewRBC(id, n, f, instanceID int) *RBC {
	return &RBC{id: id, n: n, f: f, instanceID: instanceID, State: NewState(n)}
}

func (r *RBC) SetLeader() {
	r.leader = true
}

func (r *RBC) Send(data []byte) ([]Message, error) {
	if !r.leader {
		return nil, errors.New("not leader cannot send")
	}
	msgs := make([]Message, r.n)
	for i := 0; i < r.n; i++ {
		msgs[i] = Message{fromID: r.id, destID: i, mtype: SEND, data: data}
	}
	return msgs, nil
}

func (r *RBC) Recv(m Message) ([]Message, error) {
	if r.instanceID != m.instanceID {
		return nil, errors.New("wrong instanceID")
	}
	if m.destID != r.id {
		return nil, errors.New("wrong receiver id")
	}
	log.Printf("node %v receieve %v from node %v", r.id, m.mtype, m.fromID)
	var msgs []Message
	switch m.mtype {
	case SEND:
		if r.Data != nil {
			return nil, errors.New("duplicate send message")
		}
		r.Data = m.data
		for i := 0; i < r.n; i++ {
			msg := Message{fromID: r.id, destID: i, data: hasher.SHA256Hasher(r.Data), mtype: ECHO, index: m.index}
			msgs = append(msgs, msg)
		}
	case ECHO:

		if r.Data == nil {
			return nil, errors.New("invalid echo message, no send")
		}
		if !r.nEchos[m.fromID] {
			r.nEchos[m.fromID] = true
			r.Echos++
		} else {
			return nil, nil
		}

	case READY:
		if r.Data == nil {
			return nil, errors.New("invalid ready message, no send")
		}
		if !r.nReadys[m.fromID] {
			r.nReadys[m.fromID] = true
			r.Readys++
		} else {
			return nil, nil
		}
	default:
		return nil, ErrUnknownMessageType
	}
	if r.Echos >= (r.n+r.f+1)/2 {
		for i := 0; i < r.n; i++ {
			msg := Message{fromID: r.id, destID: i, data: hasher.SHA256Hasher(r.Data), mtype: READY, index: m.index}
			msgs = append(msgs, msg)
		}
	}
	if r.Readys >= r.f+1 {
		for i := 0; i < r.n; i++ {
			msg := Message{fromID: r.id, destID: i, data: hasher.SHA256Hasher(r.Data), mtype: READY, index: m.index}
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
func NewState(n int) *State {
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
