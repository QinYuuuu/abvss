package badkg

import (
	"fmt"
	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"log/slog"
	"math/big"
	"math/rand"
	"sync"
	"testing"
)

func TestACSSShare(t *testing.T) {
	// 设置测试参数
	f := int64(1)
	degree := int64(1)
	nodeNum := int64(4)
	batchSize := int64(3)
	r := int64(1)
	sessionID := int64(0)
	p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	group := edwards25519.NewBlakeSHA256Ed25519()

	// make send channels
	channels := make([]chan *protobuf.OptRBCMessage, nodeNum)
	osvChans := make([]chan *protobuf.OSVMessage, nodeNum)
	for i := range channels {
		channels[i] = make(chan *protobuf.OptRBCMessage, 100)
		osvChans[i] = make(chan *protobuf.OSVMessage, 100)
	}
	send := func(targetID int64, msg *protobuf.OptRBCMessage) {
		channels[targetID] <- msg
	}
	osvSend := func(msg *protobuf.OSVMessage) {
		osvChans[msg.DestID] <- msg
	}

	// make acss instance
	s := make([]*big.Int, batchSize)
	randSource := rand.New(rand.NewSource(0))
	for i := range s {
		s[i] = new(big.Int).Rand(randSource, p)
	}
	acssNodes := make([]*ACSSImpl, nodeNum)

	// generate key pairs
	sk := group.Scalar().SetInt64(555)
	pk := group.Point().Mul(sk, nil)

	for i := range acssNodes {
		if i == 0 {
			acssNodes[0] = NewACSSImplDealer(0, degree, nodeNum, batchSize, r, sessionID, s, p, group)
			acssNodes[0].pkList = make([]kyber.Point, nodeNum)
			for j := int64(0); j < nodeNum; j++ {
				// use same pk
				acssNodes[0].pkList[j] = pk
				acssNodes[0].sk = sk
			}
			continue
		}
		acssNodes[i] = NewACSSImpl(int64(i), degree, nodeNum, batchSize, r, sessionID, 0, p, group)
		acssNodes[i].pkList = make([]kyber.Point, nodeNum)
		for j := int64(0); j < nodeNum; j++ {
			// use same pk
			acssNodes[i].pkList[j] = pk
			acssNodes[i].sk = sk
		}
	}

	// set RBC
	for i := int64(0); i < nodeNum; i++ {
		pid := i
		recvFunc := func() (*protobuf.OptRBCMessage, bool) {
			select {
			case msg := <-channels[pid]:
				return msg, true
			default:
				return nil, false
			}
		}
		osvRecvFunc := func() chan *protobuf.OSVMessage {
			return osvChans[pid]
		}
		acssNodes[i].rbc = broadcast.NewOptRBC(i, nodeNum, f, send, recvFunc)
		acssNodes[i].osvNode = osv.NewInstance(nodeNum, f, i, "0", osvSend, osvRecvFunc)
		acssNodes[i].Run()
	}
	// 生成并验证共享
	acssNodes[0].Share() // 为节点2生成共享\
	// outputShare := make([]*protobuf.SS24Share, nodeNum)
	var wg sync.WaitGroup
	wg.Add(4)
	for i := range nodeNum {
		select {
		case output := <-acssNodes[i].output:
			slog.Info(fmt.Sprintf("[node %v] %v", i, output.FShare))
			wg.Done()
		}
	}
	wg.Wait()
}
