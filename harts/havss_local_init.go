package harts

import (
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
)

func InitLocalMultiHAVSS(
	n, tc, tr, dealerID int64,
	instanceID string,
	group kyber.Group,
	nizkIPAParam *nizk.NizkIPAParam,
	pedersenParam *pedersen.VectorParam,
	rbcList []*broadcast.OptRBC,
) []*HAVSSImpl {
	// init 4 havssMsgChans
	havssMsgChans := make([]chan *protobuf.HartsHavssMessage, n)
	sendHAVSSMsg := func(msg *protobuf.HartsHavssMessage) {
		havssMsgChans[msg.DestID] <- msg
	}

	// init 4 HAVSS implementation
	havss := make([]*HAVSSImpl, n)

	for i := int64(0); i < n; i++ {
		havssMsgChans[i] = make(chan *protobuf.HartsHavssMessage, n*100)
		recvFunc := func() chan *protobuf.HartsHavssMessage {
			return havssMsgChans[i]
		}
		havssNetwork := HAVSSNetwork{
			bandwidthCounter: 0,
			send:             sendHAVSSMsg,
			receive:          recvFunc,
		}
		if i == dealerID {
			havss[i] = NewHAVSSDealerImpl(i, n, tc, tr, instanceID, group, nizkIPAParam, pedersenParam, havssNetwork)
		} else {
			havss[i] = NewHAVSSImpl(i, n, tc, tr, dealerID, instanceID, group, nizkIPAParam, pedersenParam, havssNetwork)
		}
		havss[i].rbc = rbcList[i]
	}
	return havss
}
