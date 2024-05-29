package protocol

import (
	"crypto/elliptic"
	"testing"

	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/stretchr/testify/assert"
)

func TestNewVSS(t *testing.T) {
	N := 4
	f := 1
	batchsize := 2
	vnum := 1

	var c curve.Curve
	c = elliptic.P224()
	param := c.Params()
	p := param.P
	sk := make([]*paillier.PrivateKey, N)
	pk := make([]*paillier.PublicKey, N)
	vss := make([]*ABVSS, N)
	for i := 0; i < N; i++ {
		ski, pki, err := paillier.KeyGen()
		assert.Nil(t, err, "err in KeyGen")
		sk[i] = ski
		pk[i] = pki
	}
	for i := 0; i < n; i++ {
		vss[i] = NewVSS(i, N, f, batchsize, vnum, p)
	}
}
