package pedersen

import (
	"math/big"

	"go.dedis.ch/kyber/v3"
)

var zeroBig = new(big.Int).SetInt64(0)
var oneBig = new(big.Int).SetInt64(1)

type VectorParam struct {
	group kyber.Group
	g     []kyber.Point
}

/*
Vector Pedersen Commitment
Given an array of values, we commit the array with different generators
for each element.
*/
func VectorPCommit(param VectorParam, value []kyber.Scalar) kyber.Point {
	commitment := param.group.Point().Mul(param.group.Scalar().SetInt64(0), nil)
	for i := 0; i < len(value); i++ {
		// mGs
		mG := param.group.Point().Mul(value[i], param.g[i])
		commitment = param.group.Point().Add(commitment, mG)
	}
	return commitment
}
