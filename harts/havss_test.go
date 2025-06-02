package harts

import (
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/crypto/zk/nizk"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3/util/random"
	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func TestHAVSS(t *testing.T) {
	n := int64(4)
	tc := int64(1)
	tr := int64(1)
	// p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	group := edwards25519.NewBlakeSHA256Ed25519()
	rand := random.New()
	nizkIPAParam := nizk.SetupNizkIPA(group, tc+1, rand)
	pedersenParam := pedersen.NewVectorParamWithG(group, nizkIPAParam.GetCRS().GetG())
	// init 4 HAVSS implementation
	havss := make([]*HAVSSImpl, 4)
	havss[0] = NewHAVSSDealerImpl(0, n, tc, tr, "test", group, nizkIPAParam, pedersenParam)
	for i := int64(1); i < n; i++ {
		havss[i] = NewHAVSSImpl(i, n, tc, 3, 0, "test", group, nizkIPAParam, pedersenParam)
	}

	// init 4 havssMsgChans
	havssMsgChans := make([]chan *protobuf.HartsHavssMessage, 4)
	sendHAVSSMsg := func(msg *protobuf.HartsHavssMessage) {
		havssMsgChans[msg.DestID] <- msg
	}
	for i := int64(0); i < n; i++ {
		havssMsgChans[i] = make(chan *protobuf.HartsHavssMessage, 10)
		havss[i].send = sendHAVSSMsg
		havss[i].receive = func() chan *protobuf.HartsHavssMessage {
			return havssMsgChans[i]
		}
	}
	rbcMsgChans := make([]chan *protobuf.OptRBCMessage, 4)
	sendRBCMsg := func(destID int64, msg *protobuf.OptRBCMessage) {
		rbcMsgChans[msg.DestID] <- msg
	}
	for i := int64(0); i < n; i++ {
		rbcMsgChans[i] = make(chan *protobuf.OptRBCMessage, 10)
		receiveFunc := func() (*protobuf.OptRBCMessage, bool) {
			msg, ok := <-rbcMsgChans[i]
			return msg, ok
		}
		havss[i].rbc = broadcast.NewOptRBC(i, n, tc, sendRBCMsg, receiveFunc)
	}

	// start 4 havss
	for i := int64(0); i < n; i++ {
		havss[i].Run()
	}
	havss[0].CommitAndDistribute()

	var wg sync.WaitGroup
	wg.Add(int(n))
	for i := int64(0); i < n; i++ {
		go func(i int64) {
			_ = havss[i].Output()
			wg.Done()
		}(i)
	}
	wg.Wait()
}
