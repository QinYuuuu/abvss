package pedersen

import (
	"math/big"

	"go.dedis.ch/kyber/v4"
)

var zeroBig = new(big.Int).SetInt64(0)
var oneBig = new(big.Int).SetInt64(1)

type VectorParam struct {
	group kyber.Group
	g     []kyber.Point
}

func NewVectorParam(group kyber.Group, n int64) VectorParam {
	g := make([]kyber.Point, n)
	for i := int64(0); i < n; i++ {
		g[i] = group.Point().Base()
	}
	return VectorParam{
		group: group,
		g:     g,
	}
}

func NewVectorParamWithG(group kyber.Group, g []kyber.Point) *VectorParam {
	return &VectorParam{
		group: group,
		g:     g,
	}
}

/*
Commit implement Vector Pedersen Commitment
Given an array of values, we commit the array with different generators
for each element.
*/
func (param *VectorParam) Commit(value []kyber.Scalar) kyber.Point {
	commitment := param.group.Point().Mul(param.group.Scalar().SetInt64(0), nil)
	for i := 0; i < len(value); i++ {
		// mGs
		mG := param.group.Point().Mul(value[i], param.g[i])
		commitment = param.group.Point().Add(commitment, mG)
	}
	return commitment
}

func (param *VectorParam) Open(commitment kyber.Point, value []kyber.Scalar) bool {
	return true
}
