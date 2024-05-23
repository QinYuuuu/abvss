package batchnizk

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/paillier"
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
