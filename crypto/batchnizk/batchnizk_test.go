package batchnizk

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"testing"

	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/stretchr/testify/assert"
)

func TestNewBatchNIZK(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	pk := paillier.PublicKey{N: param.P}
	zk := NewBatchNIZK(c, generator, param.P, 2, pk)
	fmt.Println(zk)
}

func TestBatchNIZK(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	batchsize := 1
	//degree := 1
	//randstate := rand.New(rand.NewSource(1))
	pk := paillier.PublicKey{N: param.P}
	zk := NewBatchNIZK(c, generator, param.P, batchsize, pk)
	/*
		secret := make([]*big.Int, batchsize)
		polyf := make([]polynomial.Polynomial, batchsize)
		for i := 0; i < batchsize; i++ {
			secret[i] = new(big.Int).SetInt64(rand.Int63())
			poly, err := polynomial.NewRand(degree, randstate, param.P)
			assert.Nil(t, err, "err in NewRand")
			err = poly.SetCoefficientBig(0, secret[i])
			assert.Nil(t, err, "err in SetCoefficientBig")
			polyf[i] = poly
		}
	*/
	var err error
	fij := make([]*big.Int, batchsize)
	zij := make([]*big.Int, batchsize)
	rij := make([]*big.Int, batchsize)
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		//fij[i] = utils.RandomNum(param.P)
		fij[i] = new(big.Int).SetInt64(1)
		rij[i] = new(big.Int).SetInt64(1)
		//zij[i], rij[i], err = pk.Encrypt(fij[i])
		zij[i], err = pk.EncryptWithR(fij[i], rij[i])
		assert.Nil(t, err, "err in Paillier Encrypt")
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	pi, err := zk.Prove(fij, rij)
	assert.Nil(t, err, "err in nizk proof")
	result, err := zk.Verify(Aijx, Aijy, zij, pi)
	assert.Nil(t, err, "err in nizk verify")
	assert.Equal(t, true, result, "zk verify")
}
