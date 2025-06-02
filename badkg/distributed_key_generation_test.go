package badkg

import (
	"fmt"
	"log/slog"
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/QinYuuuu/abvss/broadcast"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func TestDKGImpl_Run(t *testing.T) {
	f := int64(1)
	degree := int64(1)
	nodeNum := int64(4)
	batchSize := int64(3)
	p, _ := new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)
	group := edwards25519.NewBlakeSHA256Ed25519()

	// make send channels
	broadcastChans := make([]map[string]chan *protobuf.OptRBCMessage, nodeNum)
	osvChans := make([]map[string]chan *protobuf.OSVMessage, nodeNum)
	dkgChans := make([]chan *protobuf.DKGMessage, nodeNum)
	for i := range broadcastChans {
		broadcastChans[i] = make(map[string]chan *protobuf.OptRBCMessage)
		osvChans[i] = make(map[string]chan *protobuf.OSVMessage)
		var j int64
		for j = 0; j < nodeNum; j++ {
			broadcastChans[i][strconv.FormatInt(j, 10)+"0"] = make(chan *protobuf.OptRBCMessage, 10)
			broadcastChans[i][strconv.FormatInt(j, 10)+"1"] = make(chan *protobuf.OptRBCMessage, 10)
			broadcastChans[i][strconv.FormatInt(j, 10)+"2"] = make(chan *protobuf.OptRBCMessage, 10)
			broadcastChans[i][strconv.FormatInt(j, 10)+"3"] = make(chan *protobuf.OptRBCMessage, 10)

			osvChans[i][strconv.FormatInt(j, 10)] = make(chan *protobuf.OSVMessage, 10)
		}

	}
	rbcSend := func(targetID int64, msg *protobuf.OptRBCMessage) {
		if targetID != 0 {
			fmt.Print("")
		}
		broadcastChans[targetID][msg.SessionID] <- msg
	}

	osvSend := func(msg *protobuf.OSVMessage) {
		osvChans[msg.DestID][msg.InstanceID] <- msg
	}
	dkgSend := func(msg *protobuf.DKGMessage) {
		dkgChans[msg.DestID] <- msg
	}

	// generate key pairs
	sk := group.Scalar().SetInt64(555)
	pk := group.Point().Mul(sk, nil)

	dkgNodes := make([]*DKGImpl, nodeNum)
	dkgOutput := make(chan []*protobuf.SS24Share)
	var i int64
	for i = 0; i < nodeNum; i++ {
		id := i
		recChan := func() chan *protobuf.DKGMessage {
			return dkgChans[id]
		}
		dkgNodes[i] = NewDKGImpl(i, degree, nodeNum, batchSize, p, group, nil, nil, nil, dkgOutput, dkgSend, recChan)
		var j int64
		for j = 0; j < nodeNum; j++ {
			acssID := j
			recvFunc := func() (*protobuf.OptRBCMessage, bool) {
				select {
				case msg := <-broadcastChans[id][strconv.FormatInt(acssID, 10)+"0"]:
					return msg, true
				case msg := <-broadcastChans[id][strconv.FormatInt(acssID, 10)+"1"]:
					return msg, true
				case msg := <-broadcastChans[id][strconv.FormatInt(acssID, 10)+"2"]:
					return msg, true
				case msg := <-broadcastChans[id][strconv.FormatInt(acssID, 10)+"3"]:
					return msg, true
				default:
					return nil, false
				}
			}
			recvOSVChan := func() chan *protobuf.OSVMessage {
				osvChan := osvChans[id][strconv.FormatInt(acssID, 10)]
				return osvChan
			}
			dkgNodes[i].acssImpls[j].rbc = broadcast.NewOptRBC(i, nodeNum, f, rbcSend, recvFunc)
			dkgNodes[i].acssImpls[j].osvNode = osv.NewInstance(nodeNum, f, i, strconv.FormatInt(j, 10), osvSend, recvOSVChan)
			// use same pk
			dkgNodes[i].acssImpls[j].pkList = make([]kyber.Point, nodeNum)
			for l := 0; l < int(nodeNum); l++ {
				dkgNodes[i].acssImpls[j].pkList[l] = pk
			}
			dkgNodes[i].acssImpls[j].sk = sk
		}
		go dkgNodes[i].Run()
	}

	var wg sync.WaitGroup
	//done := make(chan bool)
	wg.Add(int(nodeNum))
	//go func() {
	for i = 0; i < nodeNum; i++ {
		id := i
		go func(id int64) {
			defer wg.Done()
			output := <-dkgNodes[id].output
			slog.Info("output", slog.Any("id", i), slog.Any("output", output))
		}(id)
	}
	wg.Wait()
	//done <- true
	//}()
	/*
		select {
		case <-done:
			return
		case <-time.After(30 * time.Second):
			t.Error("测试超时，10秒后强制终止")
			return
		}*/
}
