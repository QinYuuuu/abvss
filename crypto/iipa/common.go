package iipa

import (
	"crypto/cipher"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"go.dedis.ch/kyber/v3"
	"math/big"
)

type CRS struct {
	n     int64
	p     *big.Int
	g     []kyber.Point
	h     kyber.Point
	y     []*big.Int
	group kyber.Group
	r     cipher.Stream
}

func NewCRS(n int64, group kyber.Group, r cipher.Stream, p *big.Int) *CRS {
	g := make([]kyber.Point, n)
	y := make([]*big.Int, n)
	var i int64
	for i = 0; i < n; i++ {
		exp := group.Scalar().Pick(r)
		g[i] = group.Point().Mul(exp, nil)
		y[i] = utils.RandomNum(p)
	}
	exp := group.Scalar().Pick(r)
	h := group.Point().Mul(exp, nil)
	return &CRS{
		n:     n,
		p:     p,
		g:     g,
		h:     h,
		y:     y,
		group: group,
		r:     r,
	}
}
