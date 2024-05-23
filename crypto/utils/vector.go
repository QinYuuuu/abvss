package utils

import (
	"errors"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/curve"
)

func DotProduct(v1, v2 []*big.Int) (*big.Int, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}
	dot := zero
	for i := 0; i < len(v1); i++ {
		dot.Add(dot, new(big.Int).Mul(v1[i], v2[i]))
	}
	return dot, nil
}

func DotProductGroup(curve curve.Curve, v1 []*big.Int, v2x, v2y []*big.Int) (*big.Int, *big.Int, error) {
	if len(v1) != len(v2x) || len(v1) != len(v2y) || len(v2x) != len(v2y) {
		return nil, nil, errors.New("the input length is different")
	}
	dotx := zero
	doty := zero
	for i := 0; i < len(v1); i++ {
		tmpx, tmpy := curve.ScalarMult(v2x[i], v2y[i], v1[i].Bytes())
		curve.Add(dotx, doty, tmpx, tmpy)
	}
	return dotx, doty, nil
}

func VecPow(v1, v2 []*big.Int) (*big.Int, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}
	dot := one
	for i := 0; i < len(v1); i++ {
		dot.Add(dot, new(big.Int).Mul(v1[i], v2[i]))
	}
	return dot, nil
}
