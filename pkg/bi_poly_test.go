package pkg

import (
	"math/big"
	"testing"
)

func TestBiPolyEvalAtXMod(t *testing.T) {
	p := big.NewInt(17)
	poly := &BivariatePoly{
		coeff: [][]*big.Int{
			{big.NewInt(3), big.NewInt(1)}, // 3*y^0 + 1*y^1
			{big.NewInt(2), big.NewInt(4)}, // 2*x^1*y^0 +4*x^1*y^1
		},
		degreeX: 1,
		degreeY: 1,
	}

	// 计算在x=2处的y多项式
	yPoly := poly.evaluatePolynomialAtX(2)
}
