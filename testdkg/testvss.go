package main

import (
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"github.com/QinYuuuu/abvss/internal/abvss"
	"github.com/QinYuuuu/abvss/internal/osv"
	"github.com/QinYuuuu/abvss/network"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
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
