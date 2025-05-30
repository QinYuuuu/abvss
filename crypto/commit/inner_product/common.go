package inner_product

import (
	"crypto/cipher"

	"go.dedis.ch/kyber/v4"
)

type CRS struct {
	n     int64
	gVec  []kyber.Point
	h     kyber.Point
	group kyber.Group
	r     cipher.Stream
}

func (crs *CRS) GetN() int64 {
	return crs.n
}

func (crs *CRS) GetG() []kyber.Point {
	return crs.gVec
}

func (crs *CRS) GetH() kyber.Point {
	return crs.h
}

func (crs *CRS) GetGroup() kyber.Group {
	return crs.group
}

func (crs *CRS) GetRand() cipher.Stream {
	return crs.r
}

func NewCRS(n int64, group kyber.Group, rand cipher.Stream) *CRS {
	g := make([]kyber.Point, n)
	y := make([]kyber.Scalar, n)
	var i int64
	for i = 0; i < n; i++ {
		g[i] = group.Point().Pick(rand)
		y[i] = group.Scalar().Pick(rand)
	}
	h := group.Point().Pick(rand)
	return &CRS{
		n:     n,
		gVec:  g,
		h:     h,
		group: group,
		r:     rand,
	}
}

type Proof struct {
	lVec       []kyber.Point
	rVec       []kyber.Point
	aVecToSend []kyber.Scalar
}

type Commitment struct {
	A kyber.Point
}
