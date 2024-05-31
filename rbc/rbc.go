package rbc

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/hasher"
)

const SEND = "S"
const ECHO = "E"
const READY = "R"

var ErrUnknownMessageType = errors.New("unknown protocol message type")

type RBC struct {
	id, n, f int
	out      chan []byte
	count    int
	stateMap map[int]*State
}

type State struct {
	id, n, f        int
	Data            []byte
	Echos, Readys   int
	nEchos, nReadys []bool
	SentReady       bool
	Output          bool
}
type Message struct {
	fromID, destID, index int
	mtype                 string
	data                  []byte
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

func (s *State) Recv(m Message) ([]Message, error) {
	var msgs []Message
	switch m.mtype {
	case SEND:
		if s.Data != nil {
			return nil, errors.New("duplicate send message")
		}
		s.Data = m.data
		for i := 0; i < s.n; i++ {
			msg := Message{fromID: s.id, destID: i, data: hasher.SHA256Hasher(s.Data), mtype: ECHO, index: m.index}
			msgs = append(msgs, msg)
		}
	case ECHO:

		if s.Data == nil {
			return nil, errors.New("invalid echo message, no send")
		}
		if !s.nEchos[m.fromID] {
			s.nEchos[m.fromID] = true
			s.Echos++
		} else {
			return nil, errors.New("duplicate echo message")
		}

	case READY:
		if s.Data == nil {
			return nil, errors.New("invalid ready message, no send")
		}
		if !s.nReadys[m.fromID] {
			s.nReadys[m.fromID] = true
			s.Readys++
		} else {
			return nil, errors.New("duplicate ready message")
		}
	default:
		return nil, ErrUnknownMessageType
	}
	if s.Echos >= (s.n+s.f+1)/2 {
		for i := 0; i < s.n; i++ {
			msg := Message{fromID: s.id, destID: i, data: hasher.SHA256Hasher(s.Data), mtype: READY, index: m.index}
			msgs = append(msgs, msg)
		}
	}
	if s.Readys >= s.f+1 {
		for i := 0; i < s.n; i++ {
			msg := Message{fromID: s.id, destID: i, data: hasher.SHA256Hasher(s.Data), mtype: READY, index: m.index}
			msgs = append(msgs, msg)
		}
	}
	if s.Readys >= 2*s.f+1 {
		s.Output = true
	}
	return msgs, nil
}
