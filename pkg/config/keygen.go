package config

import (
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
	"os"
)

type ElgamalKeyPair struct {
	SK kyber.Scalar
	PK kyber.Point
}

func ElgamalCurve25519KeyGen(n int, addr string) {
	suite := curve25519.NewBlakeSHA256Curve25519(true)
	keypairs := make([]ElgamalKeyPair, n)
	for i := 0; i < n; i++ {
		pk, sk := elgamal.KeyGenCurve25519(suite)
		keypairs[i] = ElgamalKeyPair{
			PK: pk,
			SK: sk,
		}
	}
	f, err := os.Open(addr + "keypairs")
}
