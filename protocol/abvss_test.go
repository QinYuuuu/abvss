package protocol

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"testing"

	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
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
	pk := make([]PublicKey, N)
	sk := make([]SecretKey, N)
	vss := make([]*ABVSS, N)
	for i := 0; i < N; i++ {
		ski, pki, err := paillier.KeyGen()
		assert.Nil(t, err, "err in KeyGen")
		sk[i] = &paillierSecretKey{ski}
		pk[i] = &paillierPubKey{pki}
	}
	for i := 0; i < N; i++ {
		vssi, err := NewVSS(0, i, N, f, batchsize, vnum, p)
		assert.Nil(t, err, "err in NewVSSS")
		vss[i] = vssi
	}
}

func TestABVSSD(t *testing.T) {
	N := 4
	f := 1
	batchsize := 2
	vnum := 1

	var c curve.Curve
	c = elliptic.P224()
	param := c.Params()
	p := param.P
	pk := make([]PublicKey, N)
	sk := make([]SecretKey, N)
	vss := make([]*ABVSS, N)
	for i := 0; i < N; i++ {
		ski, pki, err := paillier.KeyGen()
		assert.Nil(t, err, "err in KeyGen")
		sk[i] = &paillierSecretKey{ski}
		pk[i] = &paillierPubKey{pki}
	}
	for i := 0; i < N; i++ {
		vssi, err := NewVSS(0, i, N, f, batchsize, vnum, p)
		assert.Nil(t, err, "err in NewVSSS")
		vss[i] = vssi
	}
	s := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		s[i] = utils.RandomNum(p)
	}
	fmt.Printf("secrets %v\n", s)
	err := vss[0].DistributorInit(pk, s)
	assert.Nil(t, err, "err in DistributorInit")
	err = vss[0].SamplePoly()
	assert.Nil(t, err, "err in SamplePoly")
	fmt.Printf("polyf %v\n", vss[0].polyf)
	fmt.Printf("polyg %v\n", vss[0].polyf)
	zi := make([][]Cipher, N)
	for i := 0; i < N; i++ {
		zii, _, err := vss[0].GenerateShares(i)
		assert.Nil(t, err, "err in GenerateShares")
		zi[i] = zii
	}
	xlist := make([]*big.Int, N)
	shares := make([]*big.Int, N)
	for i := 0; i < N; i++ {
		xlist[i] = new(big.Int).SetInt64(int64(i + 1))
		shares[i], _ = sk[i].Decrypt(zi[i][0])
	}
	poly, err := polynomial.LagrangeInterpolation(xlist, shares, p)
	assert.Nil(t, err, "err in LagrangeInterpolation")
	getsecret, _ := poly.GetCoefficient(0)
	assert.Equal(t, s[0], getsecret, "secret from LagrangeInterpolation")
}
