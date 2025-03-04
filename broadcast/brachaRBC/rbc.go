package brachaRBC

import (
	"log"
	"sync"
)

type Message struct {
	FromID int
	DestID int
	Mtype  string
}

// BroadCast
/*
Implementation of Validated Reliable Broadcast from DXL21 with good case optimization.
Briefly, the protocol proceeds as follows:
1. Broadcaster sends the proposal to all
2. Nodes run Bracha's RBC on hash
3. Node i output once the RBC on hash terminates and if it has received a matching proposal from leader
4. Otherwise, node i triggers a fallback protocol that uses ADD to help node i recover the proposal.
*/
func BroadCast(fromID int, destID int, mtype string, message []byte, broadcast func(id int, message *Message)) {

}

func (n *Node) handlePropose(msg Message) {
	senderID := 0 // TODO: 需要从消息中获取真实发送者ID
	if senderID != n.Leader {
		log.Printf("[%d] PROPOSE message from other than leader: %d", n.ID, senderID)
		return
	}

	leaderMsg := msg.Payload
	if n.Predicate(leaderMsg) {
		n.leaderHash = hash(leaderMsg)
		n.Broadcast(Message{ECHO, n.leaderHash})

		if n.leaderHash == n.committedHash {
			n.Broadcast(Message{TERMINATE, nil})
		}
	}
}
