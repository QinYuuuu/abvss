package osv

import (
	"errors"
	"log"
)

const Echo string = "osv.Echo"
const Vote string = "osv.V"

type Message struct {
	FromID     int
	DestID     int
	InstanceID int
	Mtype      string
}

func (m Message) Dest() int {
	return m.DestID
}

func (m Message) From() int {
	return m.FromID
}

func (m Message) Type() string {
	return m.Mtype
}

type OSV struct {
	n, t, id int
	echosNum int
	votesNum int
	nEchos   []bool
	nVotes   []bool
	acquired bool
	voted    bool
	done     bool
	OutPut   chan bool
}

func NewOSV(n, t, id int) *OSV {
	return &OSV{
		n:        n,
		t:        t,
		id:       id,
		echosNum: 0,
		votesNum: 0,
		nEchos:   make([]bool, n),
		nVotes:   make([]bool, n),
		acquired: false,
		voted:    false,
		done:     false,
		OutPut:   make(chan bool),
	}
}

func (osv *OSV) Init() []Message {
	if osv.acquired {
		log.Printf("node has acquired")
	}
	//log.Printf("node %v osv init", osv.id)
	var msgs []Message
	for i := 0; i < osv.n; i++ {
		msg := Message{}
		msg.FromID = osv.id
		msg.DestID = i
		msg.Mtype = Echo
		if i == osv.id {
			newMsgs := osv.Loop(msg)
			msgs = append(msgs, newMsgs...)
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (osv *OSV) Done() bool {
	return osv.done
}

func (osv *OSV) Loop(m Message) []Message {
	msgs, _ := osv.Recv(m)
	i := 0
	flag := len(msgs)
	for i < flag {
		if msgs[i].Dest() == osv.id {
			newMsgs, _ := osv.Recv(msgs[i])
			msgs = append(msgs[:i], newMsgs...)
			i = 0
			flag = len(msgs)
		}
		i++
	}
	return msgs
}

func (osv *OSV) Recv(m Message) ([]Message, error) {
	var msgs []Message
	if m.DestID != osv.id {
		return nil, errors.New("wrong destination id")
	}
	if m.Mtype == Echo {
		//osv.handleEcho(m)
		//log.Printf("[node %v] received ECHO from node %v", osv.id, m.FromID)
		osv.echosNum += 1
	}
	if m.Mtype == Vote {
		//osv.handleVote(m)
		if osv.nVotes[m.FromID] {
			//log.Printf("node %v has already voted", m.fromID)
			return nil, nil
		}
		//log.Printf("[node %v] received VOTE from node %v", osv.id, m.FromID)
		osv.votesNum += 1
		osv.nVotes[m.FromID] = true
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.FromID = osv.id
			msg.DestID = i
			msg.Mtype = Vote
			if i == osv.id {
				newMsgs := osv.Loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.t+1 && !osv.voted {
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.FromID = osv.id
			msg.DestID = i
			msg.Mtype = Vote
			if i == osv.id {
				newMsgs := osv.Loop(msg)
				msgs = append(msgs, newMsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.n-osv.t && osv.voted {
		if osv.done == false {
			osv.done = true
			osv.OutPut <- true
		}
		//log.Printf("[node %v] output %v", osv.id, osv.done)
		return nil, nil
	}
	return nil, nil
}
