package batchnizk

import (
	"crypto/elliptic"
	"fmt"
	"math/big"
	"testing"

	"github.com/QinYuuuu/abvss/crypto/utils"

	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/stretchr/testify/assert"
)

/*
	func TestNewBatchNIZK(t *testing.T) {
		var c curve.Curve
		c = elliptic.P224()
		param := c.Params()
		generator := curve.NewECPoint(param.Gx, param.Gy)
		pk := paillier.PublicKey{N: param.P}
		zk := NewBatchNIZK(c, generator, 2, &pk)
		fmt.Println(zk, param.P, param.Gx, param.Gy)
		fmt.Print(c.ScalarBaseMult(param.N.Bytes()))
	}
*/
func TestBatchNIZK(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	batchsize := 3
	//n := new(big.Int).Mul(param.P, big.NewInt(5))
	//degree := 1
	//randstate := rand.New(rand.NewSource(1))
	sk, pk, _ := paillier.NewKeyPair()
	/*
		n := new(big.Int).SetInt64(77)
		pk := &paillier.PublicKey{N: n}*/
	zk := NewBatchNIZK(c, generator, batchsize, pk, sk)

	var err error
	fij := make([]*big.Int, batchsize)
	zij := make([]*big.Int, batchsize)
	rij := make([]*big.Int, batchsize)
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		fij[i] = utils.RandomNum(param.P)
		//fij[i] = new(big.Int).SetInt64(1)
		rij[i] = new(big.Int).SetInt64(3)
		//rij[i], err = utils.RandomPrimeNum(pk.N)
		assert.Nil(t, err, "err in RandomPrimeNum")
		zij[i], err = pk.EncryptWithR(fij[i], rij[i])
		//zij[i], rij[i], err = pk.Encrypt(fij[i])
		fmt.Printf("zij:\t%v\n", zij[i])
		fmt.Printf("rij:\t%v\n", rij[i])

		assert.Nil(t, err, "err in Paillier Encrypt")
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	pi, err := zk.Prove(fij, rij)
	assert.Nil(t, err, "err in nizk proof")
	result, err := zk.Verify(Aijx, Aijy, zij, pi)
	assert.Nil(t, err, "err in nizk verify")
	assert.Equal(t, true, result, "zk verify")
}
