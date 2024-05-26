package main

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/zkp"
	"math/big"
)

func main() {
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
	zk := zkp.NewBatchNIZK(c, generator, batchsize, pk)

	fij := make([]*big.Int, batchsize)
	zij := make([]*big.Int, batchsize)
	rij := make([]*big.Int, batchsize)
	Aijx := make([]*big.Int, batchsize)
	Aijy := make([]*big.Int, batchsize)
	for i := 0; i < batchsize; i++ {
		fij[i] = utils.RandomNum(param.P)
		zij[i], rij[i], _ = pk.Encrypt(fij[i])
		Aijx[i], Aijy[i] = c.ScalarMult(generator.X(), generator.Y(), fij[i].Bytes())
	}
	pi, _ := zk.Prove(fij, rij)
	result, _ := zk.Verify(Aijx, Aijy, zij, pi)
	fmt.Println(result)
}
