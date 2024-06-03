package services

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/protobuf"
	"log"
	"math/big"
	"time"
)

type ABVSSNode struct {
	*ABVSSService
	*OSVService
	*network.Peer
}

func GenerateIplist(n int) []string {
	iplist := make([]string, n)
	for i := 0; i < n; i++ {
		iplist[i] = fmt.Sprintf("127.0.0.1:%d", 8000+i)
	}
	return iplist
}

func TestVSS() {
	n := 4
	f := 1
	batchszie := 1
	vnum := 4
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	//iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002", "127.0.0.1:8003"}
	iplist := GenerateIplist(n)
	pk := make([]paillier.PublicKey, n)
	sk := make([]paillier.PrivateKey, n)
	for i := 0; i < n; i++ {
		ski, pki, _ := paillier.KeyGen()
		for pki.N.Cmp(p) == -1 {
			ski, pki, _ = paillier.KeyGen()
		}
		sk[i] = *ski
		pk[i] = *pki
	}
	nodes := make([]*ABVSSNode, n)
	for i := 0; i < n; i++ {
		abvss, err := NewVSS(0, i, n, f, batchszie, vnum, p)
		osv := NewOSV(n, f, i)
		if err != nil {
			log.Println("NewVSS err:", err)
		}
		abvss.ReceiverInit(sk[i])
		abvss.VerifyInit()
		peer, err := network.NewPeer(n, i, iplist)

		if err != nil {
			log.Println("NewPeer err:", err)
		}
		abvssservice := NewABVSSService(n)
		abvssservice.ABVSS = abvss
		protobuf.RegisterABVSSServer(peer.Server, abvssservice)
		osvservice := NewOSVService(n)
		osvservice.OSV = osv
		protobuf.RegisterOSVServer(peer.Server, osvservice)
		go peer.Serve(false)
		peer.Connect()
		for j := 0; j < n; j++ {
			if j == i {
				continue
			}
			abvssservice.Clients[j] = protobuf.NewABVSSClient(peer.Conns[j])
			osvservice.Clients[j] = protobuf.NewOSVClient(peer.Conns[j])
		}
		nodes[i] = &ABVSSNode{ABVSSService: abvssservice, OSVService: osvservice, Peer: peer}
	}
	s := make([]*big.Int, batchszie)
	for i := 0; i < batchszie; i++ {
		s[i] = utils.RandomNum(p)
	}
	//fmt.Println(nodes[0].Conns)
	//fmt.Println(nodes[0].Clients)
	start := time.Now()
	nodes[0].SecretSharing(pk, s)

	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if nodes[i].received {
					nodes[i].BroadcastLCM()
					return
				}
			}
		}(i)
	}

	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if nodes[i].count == n-f {
					nodes[i].OSVService.Init()
					return
				}
			}
		}(i)
	}
	var flag bool
	for i := 0; i < n; i++ {
		go func(i int) {
			for {
				if nodes[i].Done() {
					flag = true
					break
				}
			}
		}(i)
	}
	for flag == false {

	}
	end := time.Now()
	//fmt.Println("SUCCESS")
	fmt.Println("Time cost:", end.Sub(start))
}
