package osv

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
		acquired: false,
		voted:    false,
		done:     false,
	}
}

func (osv *OSV) Init() []Message {
	if osv.acquired {
		log.Printf("node has acquired")
	}
	msgs := make([]Message, osv.n)
	for i := 0; i < osv.n; i++ {
		msg := Message{}
		msg.fromID = osv.id
		msg.destID = i
		msg.mtype = Echo
		msgs[i] = msg
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
func (osv *OSV) Recv(m Message) ([]Message, error) {
	if m.destID != osv.id {
		return nil, errors.New("wrong destination id")
	}
	if m.mtype == Echo {
		//osv.handleEcho(m)
		log.Printf("[node %v] received ECHO from node %v", osv.id, m.fromID)
		osv.echosNum += 1
	}
	if m.mtype == Vote {
		//osv.handleVote(m)
		log.Printf("[node %v] received VOTE from node %v", osv.id, m.fromID)
		osv.votesNum += 1
	}
	if osv.echosNum >= osv.n-osv.t && !osv.voted {
		msgs := make([]Message, osv.n)
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.fromID = osv.id
			msg.destID = i
			msg.mtype = Vote
			msgs[i] = msg
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.t+1 && !osv.voted {
		msgs := make([]Message, osv.n)
		for i := 0; i < osv.n; i++ {
			msg := Message{}
			msg.fromID = osv.id
			msg.destID = i
			msg.mtype = Vote
			msgs[i] = msg
		}
		osv.voted = true
		return msgs, nil
	}
	if osv.votesNum >= osv.n-osv.t && osv.voted {
		osv.done = true
		log.Printf("[node %v] output %v", osv.id, osv.done)
		return nil, nil
	}
	return nil, nil
}
