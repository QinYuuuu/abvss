package pkg

import (
	"fmt"
	"math/big"
)

type BivariatePoly struct {
	coefficients [][]*big.Int // 双变量多项式系数 S(x,y)
	degreeX      int          // X的度数tr
	degreeY      int          // Y的度数tc
}

func GenerateBivariatePoly(s *big.Int, tr, tc int) (*BivariatePoly, error) {
	if tr < 0 || tc < 0 {
		return nil, fmt.Errorf(fmt.Sprintf("degree must be non-negative, got %d, %d", tr, tc))
	}
	coeff := make([][]*big.Int, tr+1)
	for i := 0; i < len(coeff); i++ {
		coeff[i] = make([]*big.Int, tc+1)
		for j := 0; j < len(coeff[i]); i++ {
			coeff[i][j] = big.NewInt(0)
		}
	}
	// 初始化系数矩阵并保证S(0,0)=s
	coeff[0][0] = s
	return &BivariatePoly{
		coefficients: coeff,
		degreeX:      tr,
		degreeY:      tc,
	}, nil
}

func (poly *BivariatePoly) EvalMod(x, y, p *big.Int) {

}
