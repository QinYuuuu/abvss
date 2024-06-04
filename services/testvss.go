package main

import (
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
	"log"
	"math/big"
	"time"
)

type ABVSSNode struct {
	*abvss.ABVSSService
	*osv.OSVService
	*network.Peer
}

func GenerateIplist(n int) ([]string, []string, []string) {
	iplist := make([]string, n)
	for i := 0; i < n; i++ {
		iplist[i] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
	}
	addlist := make([]string, n)
	portlist := make([]string, n)
	for i := 0; i < n; i++ {
		addlist[i] = "127.0.0.1"
		portlist[i] = fmt.Sprintf("%d", 9000+i)
	}
	return iplist, addlist, portlist
}

func GenerateElGamal(n int, addr string) {
	pk := make([]kyber.Point, n)
	sk := make([]kyber.Scalar, n)
	suite := curve25519.NewBlakeSHA256Curve25519(true)
	for i := 0; i < n; i++ {
		pki, ski := elgamal.KeyGenCurve25519(suite)
		sk[i] = ski
		pk[i] = pki
	}
}

func TestVSS(id, n, f, batchsize, vnum int, p *big.Int, pk []kyber.Point, sk kyber.Scalar) {
	iplist, _, _ := GenerateIplist(n)

	node := new(ABVSSNode)

	abvss_instance, err := abvss.NewVSS(0, id, n, f, batchsize, vnum, p, 1)
	osv_instance := osv.NewOSV(n, f, id)
	if err != nil {
		log.Println("NewVSS err:", err)
	}
	abvss_instance.ReceiverInit(sk)
	abvss_instance.VerifyInit()
	peer, err := network.NewPeer(n, id, iplist)

	if err != nil {
		log.Println("NewPeer err:", err)
	}
	abvssservice := abvss.NewABVSSService(n)
	abvssservice.ABVSS = abvss_instance
	protobuf.RegisterABVSSServer(peer.Server, abvssservice)
	osvservice := osv.NewOSVService(n)
	osvservice.OSV = osv_instance
	protobuf.RegisterOSVServer(peer.Server, osvservice)
	go peer.Serve(false)
	peer.Connect()
	for j := 0; j < n; j++ {
		if j == id {
			continue
		}
		abvssservice.Clients[j] = protobuf.NewABVSSClient(peer.Conns[j])
		osvservice.Clients[j] = protobuf.NewOSVClient(peer.Conns[j])
	}
	node = &ABVSSNode{ABVSSService: abvssservice, OSVService: osvservice, Peer: peer}

	s := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		s[i] = utils.RandomNum(p)
	}
	//fmt.Println(nodes[0].Conns)
	//fmt.Println(nodes[0].Clients)
	start := time.Now()
	if id == 1 {
		node.SecretSharing(pk, s)
	}

	go func() {
		for {
			if node.Received {
				node.BroadcastLCM()
				return
			}
		}
	}()

	go func() {
		for {
			if node.Count == n-f {
				node.OSVService.Init()
				return
			}
		}
	}()

	var flag bool

	go func() {
		for {
			if node.Done() {
				flag = true
				break
			}
		}
	}()

	for flag == false {

	}
	end := time.Now()
	//fmt.Println("SUCCESS")
	fmt.Printf("node %v Time cost: %v\n", id, end.Sub(start))
}
