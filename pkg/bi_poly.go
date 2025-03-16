package pkg

import (
	"fmt"
	"math/big"
)

type BivariatePoly struct {
	coeff   [][]*big.Int // bi-poly coefficients S(x,y)
	degreeX int          // X degree tr
	degreeY int          // Y degree tc
}

func GenerateBivariatePoly(s *big.Int, tr, tc int) (*BivariatePoly, error) {
	if tr < 0 || tc < 0 {
		return nil, fmt.Errorf("degree must be non-negative, got %d, %d", tr, tc)
	}
	coeff := make([][]*big.Int, tr+1)
	for i := 0; i < len(coeff); i++ {
		coeff[i] = make([]*big.Int, tc+1)
		for j := 0; j < len(coeff[i]); j++ {
			coeff[i][j] = big.NewInt(0)
		}
	}
	// 初始化系数矩阵并保证S(0,0)=s
	coeff[0][0] = s
	return &BivariatePoly{
		coeff:   coeff,
		degreeX: tr,
		degreeY: tc,
	}, nil
}

func (poly *BivariatePoly) EvalAtXMod(x, p *big.Int) *Poly {
	result, _ := New(poly.degreeY + 1)
	for i := 0; i <= poly.degreeY; i++ { // y degree loop
		coeffSum := new(big.Int).SetInt64(0)
		xExp := new(big.Int).SetInt64(1)
		for k := 0; k <= poly.degreeX; k++ { // x的度数遍历
			// 获取双变量项的系数 a_ij (k对应x的度，j对应y的度)
			term := new(big.Int).Set(poly.coeff[k][i])
			// 执行相乘并累加：a_kj * x^k
			term.Mul(term, xExp)
			coeffSum.Add(coeffSum, term)
		}
		result.coeff[i].Mod(coeffSum, p)
	}
	return result
}

func (poly *BivariatePoly) EvalAtYMod(x, p *big.Int) *Poly {
	result, _ := New(poly.degreeX + 1)

	for i := 0; i <= poly.degreeX; i++ { // y degree loop
		coeffSum := new(big.Int).SetInt64(0)
		yExp := new(big.Int).SetInt64(1)
		for k := 0; k <= poly.degreeY; k++ { // x的度数遍历
			// 获取双变量项的系数 a_ij (k对应x的度，j对应y的度)
			term := new(big.Int).Set(poly.coeff[i][k])
			// 执行相乘并累加：a_kj * x^k
			term.Mul(term, yExp)
			coeffSum.Add(coeffSum, term)
		}
		result.coeff[i].Mod(coeffSum, p)
	}
	return result
}
