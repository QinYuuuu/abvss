package services

import (
	"crypto/elliptic"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/network"
	"github.com/QinYuuuu/abvss/protobuf"
	"log"
	"math/big"
)

type ABVSSNode struct {
	*ABVSSService
	*network.Peer
}

func TestVSS() {
	n := 4
	f := 1
	batchszie := 2
	vnum := 4
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	iplist := []string{"127.0.0.1:8000", "127.0.0.1:8001", "127.0.0.1:8002", "127.0.0.1:8003"}

	pk := make([]PublicKey, n)
	sk := make([]SecretKey, n)
	for i := 0; i < n; i++ {
		ski, pki, _ := paillier.KeyGen()
		sk[i] = &paillierSecretKey{ski}
		pk[i] = &paillierPubKey{pki}
	}
	nodes := make([]*ABVSSNode, n)
	for i := 0; i < n; i++ {
		abvss, err := NewVSS(0, i, n, f, batchszie, vnum, p)
		peer, err := network.NewPeer(n, i, iplist)

		if err != nil {
			log.Println("NewPeer err:", err)
		}
		abvssservice := NewABVSSService(n)

		protobuf.RegisterABVSSServer(peer.Server, abvssservice)
		go peer.Serve(false)
		peer.Connect()
		for j := 0; j < n; j++ {
			abvssservice.Clients[i] = protobuf.NewABVSSClient(peer.Conns[i])
		}
		abvssservice.ABVSS = abvss
		nodes[i] = &ABVSSNode{ABVSSService: abvssservice, Peer: peer}
	}
	s := make([]*big.Int, batchszie)
	for i := 0; i < batchszie; i++ {
		s[i] = utils.RandomNum(p)
	}
	nodes[0].SecretSharing(pk, s)
	for {

	}
}
