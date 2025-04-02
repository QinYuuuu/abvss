package osv

import (
	"errors"
	"fmt"
	"log/slog"
)

const Echo string = "osv.Echo"
const Vote string = "osv.Vote"

var duplicateMessageError = errors.New("duplicate Message")
var wrongInstanceError = errors.New("wrong instanceID")
var wrongDestinationError = errors.New("wrong destinationID")

type Message struct {
	FromID     int64  `json:"from_id"`
	DestID     int64  `json:"dest_id"`
	InstanceID int64  `json:"instance_id"`
	MsgType    string `json:"msg_type"`
}

func (m Message) Dest() int64 {
	return m.DestID
}

func (m Message) From() int64 {
	return m.FromID
}

func (m Message) Type() string {
	return m.MsgType
}

type Instance struct {
	n, t, id   int64
	instanceID int64
	echosNum   int64
	votesNum   int64
	nEchos     []bool
	nVotes     []bool
	acquired   bool
	voted      bool
	done       bool
	outPut     chan bool

	send    func(Message)
	sendAll func(string)
	receive func() chan Message
}

func NewInstance(n, t, id, instanceID int64, send func(Message), sendAll func(string), receive func() chan Message) *Instance {
	return &Instance{
		n:          n,
		t:          t,
		id:         id,
		instanceID: instanceID,
		echosNum:   0,
		votesNum:   0,
		nEchos:     make([]bool, n),
		nVotes:     make([]bool, n),
		acquired:   false,
		voted:      false,
		done:       false,
		outPut:     make(chan bool),

		send:    send,
		sendAll: sendAll,
		receive: receive,
	}
}

func (osv *Instance) init() []Message {
	if osv.acquired {
		slog.Error("node has acquired")
	}
	slog.Debug(fmt.Sprintf("node %v osv init", osv.id))
	var msgs []Message
	var i int64
	for i = 0; i < osv.n; i++ {
		msg := Message{
			FromID:     osv.id,
			DestID:     i,
			InstanceID: osv.instanceID,
			MsgType:    Echo,
		}
		if i == osv.id {
			newMsgs := osv.loop(msg)
			msgs = append(msgs, newMsgs...)
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (osv *Instance) Done() bool {
	return osv.done
}

// Run node vote in osv
func (osv *Instance) Run() {
	msgs := osv.init()
	for _, msg := range msgs {
		osv.send(msg)
	}
	for {
		select {
		case msg := <-osv.receive():
			slog.Info(fmt.Sprintf("[node %v, instance %v]", osv.id, osv.instanceID), slog.Any("msg", msg))
			newMsgs, err := osv.recv(msg)
			if err != nil {
				slog.Error("handle message", slog.Any("error", err))
				return
			}
			for _, newMsg := range newMsgs {
				osv.send(newMsg)
			}
		}
	}
}

func (osv *Instance) Output() chan bool {
	return osv.outPut
}

func (osv *Instance) loop(m Message) []Message {
	msgs, _ := osv.recv(m)
	i := 0
	flag := len(msgs)
	for i < flag {
		if msgs[i].Dest() == osv.id {
			newMsgs, _ := osv.recv(msgs[i])
			msgs = append(msgs[:i], newMsgs...)
			i = 0
			flag = len(msgs)
		}
		i++
	}
	return msgs
}

func (osv *Instance) recv(m Message) ([]Message, error) {
	var msgs []Message
	if m.DestID != osv.id {
		return nil, wrongDestinationError
	}
	if m.InstanceID != osv.instanceID {
		return nil, wrongInstanceError
	}
	if m.MsgType == Echo {
		//log.Printf("[node %v] received ECHO from node %v", osv.id, m.FromID)
		if osv.nEchos[m.FromID] {
			return nil, duplicateMessageError
		}
		osv.echosNum += 1
		osv.nEchos[m.FromID] = true
	}
	if m.MsgType == Vote {
		if osv.nVotes[m.FromID] {
			//log.Printf("node %v has already voted", m.fromID)
			return nil, duplicateMessageError
		}
		//log.Printf("[node %v] received VOTE from node %v", osv.id, m.FromID)
		osv.votesNum += 1
		osv.nVotes[m.FromID] = true
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		var i int64
		osv.voted = true
		for i = 0; i < osv.n; i++ {
			msg := Message{
				FromID:     osv.id,
				DestID:     i,
				InstanceID: osv.instanceID,
				MsgType:    Vote,
			}

			if i == osv.id {
				newMsgs := osv.loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		return msgs, nil
	}
	if osv.votesNum >= osv.t+1 && !osv.voted {
		var i int64
		osv.voted = true
		for i = 0; i < osv.n; i++ {
			msg := Message{
				FromID:     osv.id,
				DestID:     i,
				InstanceID: osv.instanceID,
				MsgType:    Vote,
			}
			if i == osv.id {
				newMsgs := osv.loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		return msgs, nil
	}
	if osv.votesNum >= osv.n-osv.t && osv.voted {
		if osv.done == false {
			osv.done = true
			osv.outPut <- true
		}
		//log.Printf("[node %v] output %v", osv.id, osv.done)
		return nil, nil
	}
	return nil, nil
}
