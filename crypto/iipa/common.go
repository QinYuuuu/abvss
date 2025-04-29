package iipa

import (
	"crypto/cipher"
	"go.dedis.ch/kyber/v3"
)

type CRS struct {
	n     int64
	g     []kyber.Point
	h     kyber.Point
	y     []kyber.Scalar
	group kyber.Group
	r     cipher.Stream
}

func NewCRS(n int64, group kyber.Group, r cipher.Stream) *CRS {
	g := make([]kyber.Point, n)
	y := make([]kyber.Scalar, n)
	var i int64
	for i = 0; i < n; i++ {
		g[i] = group.Point().Pick(r)
		y[i] = group.Scalar().Pick(r)
	}
	h := group.Point().Pick(r)
	return &CRS{
		n:     n,
		g:     g,
		h:     h,
		y:     y,
		group: group,
		r:     r,
	}
}
