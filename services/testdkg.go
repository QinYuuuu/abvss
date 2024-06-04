package main

import (
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/internal/abdkg"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/internal/party"
	"github.com/QinYuuuu/abvss/internal/smvba"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	utils2 "github.com/QinYuuuu/abvss/pkg/utils"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"google.golang.org/protobuf/proto"
	"log"
	"math/big"
	"sync"
	"time"
)

type ABDKGNode struct {
	*abdkg.ABDKGService
	*network.Peer
}

func TestDKG(id, n, f, batchsize, vnum int, pint *big.Int, pk1 []kyber.Point, sk1 kyber.Scalar, pk *share.PubPoly, sk *share.PriShare, epk kyber.Point, evk []*share.PubShare, esk *share.PriShare, testNum int, signature [][]byte) {
	iplist, ipList, portList := GenerateIplist(n)
	node := new(ABDKGNode)
	var err error
	abvss_instance := make([]*abvss.ABVSS, n)
	osv_instance := make([]*osv.OSV, n)
	for i := 0; i < n; i++ {
		abvss_instance[i], err = abvss.NewVSS(i, id, n, f, batchsize, vnum, pint, 1)
		if err != nil {
			log.Println("NewVSS err:", err)
		}
		osv_instance[i] = osv.NewOSV(n, f, id)
	}
	for i := 0; i < n; i++ {
		abvss_instance[i].ReceiverInit(sk1)
		abvss_instance[i].VerifyInit()
	}
	peer, err := network.NewPeer(n, id, iplist)
	abdkgservice := abdkg.NewABDKGService(id, n)

	abdkgservice.Vss = abvss_instance
	abdkgservice.Osv = osv_instance
	protobuf.RegisterABDKGServer(peer.Server, abdkgservice)
	go peer.Serve(false)
	peer.Connect()
	for j := 0; j < n; j++ {
		if j == id {
			continue
		}
		abdkgservice.Clients[j] = protobuf.NewABDKGClient(peer.Conns[j])
	}
	node = &ABDKGNode{ABDKGService: abdkgservice, Peer: peer}
	s := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		s[i] = utils.RandomNum(pint)
	}
	start := time.Now()
	node.SecretSharing(pk1, s)
	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if node.Vss[i].Received {
					node.BroadcastLCM(i)
					return
				}
			}
		}(i)
	}
	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if node.Vss[i].Count == n-f {
					node.Init(i)
					return
				}
			}
		}(i)
	}

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if node.Osv[i].Done() {
					wg.Done()
					return
				}
			}
		}(i)
	}
	abdkg.IIPA_Prover1(batchsize)
	wg.Wait()
	p := party.NewHonestParty(uint32(n), uint32(f), uint32(id), ipList, portList, pk, sk, epk, evk, esk)

	p.InitReceiveChannel()
	p.InitSendChannel()

	defer p.Close()
	var mu sync.Mutex
	result := make([][][]byte, testNum)
	//start := time.Now()
	for k := 0; k < testNum; k++ {
		ID := utils2.IntToBytes(0)
		pids := make([]uint32, 2*f+1)
		hashes := make([][]byte, 2*f+1)
		sigs := make([][]byte, 2*f+1)
		for i := 0; i < 2*f+1; i++ {
			pids[i] = 0
			hashes[i] = []byte("TEST")
			sigs[i] = signature[k]
		}
		value, _ := proto.Marshal(&protobuf.BLockSetValue{
			Pid:  pids,
			Hash: hashes,
		})
		validation, _ := proto.Marshal(&protobuf.BLockSetValidation{
			Sig: sigs,
		})

		wg.Add(1)

		go func(k int) {
			ans := smvba.MainProcess(p, ID, value, validation, Q)
			mu.Lock()
			result[k] = append(result[k], ans)
			mu.Unlock()
			wg.Done()

		}(k)

	}
	wg.Wait()
	end := time.Now()
	//fmt.Println("SUCCESS")
	fmt.Printf("node %v Time cost: %v\n", id, end.Sub(start))
	time.Sleep(10 * time.Second)
}
