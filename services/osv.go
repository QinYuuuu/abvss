package services

import (
	"errors"
	"log"
)

const Echo string = "E"
const Vote string = "V"

type Message struct {
	fromID int
	destID int
	mtype  string
}

func (m Message) Dest() int {
	return m.destID
}

func (m Message) From() int {
	return m.fromID
}

func (m Message) Type() string {
	return m.mtype
}

type OSV struct {
	n        int
	t        int
	id       int
	echosNum int
	votesNum int
	nVotes   []bool
	acquired bool
	voted    bool
	done     bool
}

func NewOSV(n, t, id int) *OSV {
	return &OSV{
		n:        n,
		t:        t,
		id:       id,
		echosNum: 0,
		votesNum: 0,
		nVotes:   make([]bool, n),
		acquired: false,
		voted:    false,
		done:     false,
	}
}

func (osv *OSV) Init() []Message {
	if osv.acquired {
		log.Printf("node has acquired")
	}
	log.Printf("node %v osv init", osv.id)
	var msgs []Message
	for i := 0; i < osv.n; i++ {
		msg := Message{}
		msg.fromID = osv.id
		msg.destID = i
		msg.mtype = Echo
		if i == osv.id {
			newmsgs := osv.Loop(msg)
			msgs = append(msgs, newmsgs...)
		} else {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

func (osv *OSV) Done() bool {
	return osv.done
}

/*
	func (osv *OSV) handleEcho(m Message) {
		if osv.voted {
			log.Printf("node has voted")
		}
		log.Printf("received ECHO from node %v", m.fromID)
		osv.echosNum += 1
	}

	func (osv *OSV) handleVote(m Message) {
		log.Printf("received VOTE from node %v", m.fromID)
		osv.votesNum += 1
	}
*/

func (osv *OSV) Loop(m Message) []Message {
	msgs, _ := osv.Recv(m)
	i := 0
	flag := len(msgs)
	for i < flag {
		if msgs[i].Dest() == osv.id {
			newmsgs, _ := osv.Recv(msgs[i])
			msgs = append(msgs[:i], newmsgs...)
			i = 0
			flag = len(msgs)
		}
		i++
	}
	return msgs
}

func (osv *OSV) Recv(m Message) ([]Message, error) {
	var msgs []Message
	if m.destID != osv.id {
		return nil, errors.New("wrong destination id")
	}
	if m.mtype == Echo {
		//osv.handleEcho(m)
		//log.Printf("[node %v] received ECHO from node %v", osv.id, m.fromID)
		osv.echosNum += 1
	}
	if m.mtype == Vote {
		//osv.handleVote(m)
		if osv.nVotes[m.fromID] {
			//log.Printf("node %v has already voted", m.fromID)
			return nil, nil
		}
		//log.Printf("[node %v] received VOTE from node %v", osv.id, m.fromID)
		osv.votesNum += 1
		osv.nVotes[m.fromID] = true
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.fromID = osv.id
			msg.destID = i
			msg.mtype = Vote
			if i == osv.id {
				newmsgs := osv.Loop(msg)
				msgs = append(msgs, newmsgs...)
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
			msg.fromID = osv.id
			msg.destID = i
			msg.mtype = Vote
			if i == osv.id {
				newmsgs := osv.Loop(msg)
				msgs = append(msgs, newmsgs...)
			} else {
				msgs = append(msgs, msg)
			}
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.n-osv.t && osv.voted {
		osv.done = true
		//log.Printf("[node %v] output %v", osv.id, osv.done)
		return nil, nil
	}
	return nil, nil
}
