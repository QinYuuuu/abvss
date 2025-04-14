package osv

import (
	"errors"
	"fmt"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"log/slog"
)

const Echo string = "osv.Echo"
const Vote string = "osv.Vote"

var duplicateMessageError = errors.New("duplicate Message")
var wrongInstanceError = errors.New("wrong instanceID")
var wrongDestinationError = errors.New("wrong destinationID")

type Instance struct {
	n, t, id   int64
	instanceID string
	echosNum   int64
	votesNum   int64
	nEchos     []bool
	nVotes     []bool
	acquired   bool
	voted      bool
	done       bool
	outPut     chan bool

	send    func(*protobuf.OSVMessage)
	receive func() chan *protobuf.OSVMessage
}

func NewInstance(n, t, id int64, instanceID string, send func(*protobuf.OSVMessage), receive func() chan *protobuf.OSVMessage) *Instance {
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
		receive: receive,
	}
}

func (osv *Instance) init() []*protobuf.OSVMessage {
	if osv.acquired {
		slog.Error("node has acquired")
	}
	slog.Debug(fmt.Sprintf("node %v osv init", osv.id))
	var msgs []*protobuf.OSVMessage
	var i int64
	for i = 0; i < osv.n; i++ {
		msg := &protobuf.OSVMessage{
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
	go func() {
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
	}()
}

func (osv *Instance) Output() chan bool {
	return osv.outPut
}

func (osv *Instance) loop(m *protobuf.OSVMessage) []*protobuf.OSVMessage {
	msgs, _ := osv.recv(m)
	i := 0
	flag := len(msgs)
	for i < flag {
		if msgs[i].GetDestID() == osv.id {
			newMsgs, _ := osv.recv(msgs[i])
			msgs = append(msgs[:i], newMsgs...)
			i = 0
			flag = len(msgs)
		}
		i++
	}
	return msgs
}

func (osv *Instance) recv(m *protobuf.OSVMessage) ([]*protobuf.OSVMessage, error) {
	var msgs []*protobuf.OSVMessage
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
		slog.Info(fmt.Sprintf("[node %v] received VOTE from node %v, total %v", osv.id, m.FromID, osv.votesNum))
		if osv.votesNum == 2 {
			fmt.Printf("")
		}
		osv.votesNum += 1
		osv.nVotes[m.FromID] = true
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		var i int64
		osv.voted = true
		for i = 0; i < osv.n; i++ {
			msg := &protobuf.OSVMessage{
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
			msg := &protobuf.OSVMessage{
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
			slog.Info(fmt.Sprintf("[node %v, instance %v] output", osv.id, osv.instanceID))
		}
		return nil, nil
	}
	return nil, nil
}
