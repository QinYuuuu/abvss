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
	h     []kyber.Point
}

/*
VectorPCommit implement Vector Pedersen Commitment
Given an array of values, we commit the array with different generators
for each element.
*/
func VectorPCommit(param VectorParam, value []*big.Int) kyber.Point {
	commitment := param.group.Point().Mul(param.group.Scalar().SetInt64(0), nil)
	for i := 0; i < len(value); i++ {
		// mGs
		valueScalar := param.group.Scalar().SetInt64(value[i].Int64())
		mG := param.group.Point().Mul(valueScalar, param.g[i])
		commitment = param.group.Point().Add(commitment, mG)
	}
	return commitment
}
