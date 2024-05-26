package zkp

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

func TestNewBatchNIZK(t *testing.T) {
	var c curve.Curve
	c = elliptic.P224()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	_, pk, _ := paillier.NewKeyPair()
	zk := NewBatchNIZK(c, generator, 2, pk)
	fmt.Println(zk, param.P, param.Gx, param.Gy)
	fmt.Print(c.ScalarBaseMult(param.N.Bytes()))
}

func TestBatchNIZK_Prove(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	batchsize := 1
	p := new(big.Int).SetInt64(7)
	q := new(big.Int).SetInt64(11)
	pk := &paillier.PublicKey{N: new(big.Int).Mul(p, q)}
	zk := NewBatchNIZK(c, generator, batchsize, pk)
	fij := []*big.Int{new(big.Int).SetInt64(12)}
	rij := []*big.Int{new(big.Int).SetInt64(13)}
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	getpi, err := zk.Prove(fij, rij)
	assert.Nil(t, err, "err in nizk proof")
	tmp1 := new(big.Int).Exp(pk.G(), new(big.Int).SetInt64(6), pk.N2())
	tmp2 := new(big.Int).Exp(new(big.Int).SetInt64(5), pk.N, pk.N2())
	wantpie := new(big.Int).Mod(new(big.Int).Mul(tmp1, tmp2), pk.N2())
	assert.Equal(t, getpi.e, wantpie, "zk proof e_ij")

	tmp1 = new(big.Int).Mul(fij[0], new(big.Int).SetInt64(2))
	wantpir := new(big.Int).Mod(new(big.Int).Add(new(big.Int).SetInt64(6), tmp1), pk.N)
	assert.Equal(t, wantpir, getpi.r, "zk proof R_ij")

	pow := new(big.Int).Exp(rij[0], new(big.Int).SetInt64(2), pk.N2())
	wantpiq := new(big.Int).Mod(new(big.Int).Mul(new(big.Int).SetInt64(5), pow), pk.N2())
	assert.Equal(t, wantpiq, getpi.q, "zk proof Q_ij")
}

func TestBatchNIZK_Verify(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	batchsize := 1
	_, pk, _ := paillier.KeyGen()
	zk := NewBatchNIZK(c, generator, batchsize, pk)
	fij := []*big.Int{new(big.Int).SetInt64(12)}
	rij := []*big.Int{new(big.Int).SetInt64(13)}
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	getpi, err := zk.Prove(fij, rij)
	assert.Nil(t, err, "err in nizk proof")
	zij, _ := pk.EncryptWithR(fij[0], rij[0])
	/*
		tmp1 := new(big.Int).Exp(pk.G(), fij[0], pk.N2())
		tmp2 := new(big.Int).Exp(rij[0], pk.N, pk.N2())
		wantzij := new(big.Int).Mod(new(big.Int).Mul(tmp1, tmp2), pk.N2())
		fmt.Printf("zij:%v\n", zij)
		assert.Equal(t, zij, wantzij, "encrypted zij")*/
	pow := new(big.Int).Exp(new(big.Int).SetInt64(5298), new(big.Int).SetInt64(2), pk.N2())
	wantleft := new(big.Int).Mod(new(big.Int).Mul(new(big.Int).SetInt64(3083), pow), pk.N2())
	fmt.Printf("wantleft:\t%v\n", wantleft)
	zk.Verify(Aijx, Aijy, []*big.Int{zij}, getpi)
}

func TestBatchNIZK(t *testing.T) {
	var c curve.Curve
	c = elliptic.P256()
	param := c.Params()
	generator := curve.NewECPoint(param.Gx, param.Gy)
	batchsize := 3
	//n := new(big.Int).Mul(param.P, big.NewInt(5))
	//degree := 1
	//randstate := rand.New(rand.NewSource(1))
	_, pk, _ := paillier.KeyGen()
	/*
		n := new(big.Int).SetInt64(77)
		pk := &paillier.PublicKey{N: n}*/
	zk := NewBatchNIZK(c, generator, batchsize, pk)

	var err error
	fij := make([]*big.Int, batchsize)
	zij := make([]*big.Int, batchsize)
	rij := make([]*big.Int, batchsize)
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		fij[i] = utils.RandomNum(param.P)
		//fij[i] = new(big.Int).SetInt64(1)
		//rij[i] = new(big.Int).SetInt64(3)
		//rij[i], err = utils.RandomPrimeNum(pk.N)
		//assert.Nil(t, err, "err in RandomPrimeNum")
		//zij[i], err = pk.EncryptWithR(fij[i], rij[i])
		zij[i], rij[i], err = pk.Encrypt(fij[i])
		//fmt.Printf("zij:\t%v\n", zij[i])
		//fmt.Printf("rij:\t%v\n", rij[i])

		assert.Nil(t, err, "err in Paillier Encrypt")
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	pi, err := zk.Prove(fij, rij)
	assert.Nil(t, err, "err in nizk proof")
	result, err := zk.Verify(Aijx, Aijy, zij, pi)
	assert.Nil(t, err, "err in nizk verify")
	assert.Equal(t, true, result, "zk verify")
}
