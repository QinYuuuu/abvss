package batchnizk

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"math/big"
	"testing"
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
	num := 2
	pk := paillier.PublicKey{N: param.P}
	zk := NewBatchNIZK(c, generator, param.P, num, pk)
	fij := make([]*big.Int, num)
	zij := make([]*big.Int, num)
	Aijx := make([]*big.Int, num)
	Aijy := make([]*big.Int, num)
	for i := 0; i < num; i++ {
		fij := utils.RandomNum(param.P)
	}
	zk.Prove()
}
