package harts

import (
	"fmt"
	"log/slog"
	"math/big"

	"go.dedis.ch/kyber/v3"
)

type Node struct {
	n, tc, tr int
	pid       int64
	p         *big.Int
	group     kyber.Group
	sessions  []*HAVSSNode

	receive func() (Message, bool)
	send    func()
}

type Message struct {
	FromID    int64
	DestID    int64
	MsgType   string
	SessionID int64
	Payload   []byte
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
			case Commit:
				// handle Commit

			case Row:
				// handle Row

			case Column:
				// handle ready
				s.handleColumn(msg.FromID)
			case Vote:
				// handle addTrigger
				s.handleVote(msg.FromID)
			case Done:
				s.handleDone(msg.FromID)
			default:
				panic("unhandled default case")
			}
		}
	}
}
