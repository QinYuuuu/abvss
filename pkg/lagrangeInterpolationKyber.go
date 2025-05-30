package pkg

import (
	"fmt"

	"go.dedis.ch/kyber/v4"
)

// InterpolationKyber 在KyberPoint下重新实现Lagrange插值函数
func InterpolationKyber(x []int64, y []kyber.Scalar, suite kyber.Group) (*PolyKyberImpl, error) {
	if len(x) != len(y) {
		return nil, fmt.Errorf("len(x)!=len(y): %d, %d", len(x), len(y))
	}
	if len(x) == 0 {
		return nil, fmt.Errorf("len(x) == 0")
	}
	if len(x) == 1 {
		return &PolyKyberImpl{coeff: y}, nil
	}

	degree := len(x) - 1
	resultCoeffs := make([]kyber.Scalar, degree+1)
	for i := range resultCoeffs {
		resultCoeffs[i] = suite.Scalar().Zero()
	}

	for j := 0; j < len(x); j++ {
		basis := suite.Scalar().One()
		for m := 0; m < len(x); m++ {
			if m == j {
				continue
			}
			// 原代码中 x[j] 和 x[m] 是 kyber.Scalar 类型，现在 x 是 []int64，需要转换
			xjScalar := suite.Scalar().SetInt64(x[j])
			xmScalar := suite.Scalar().SetInt64(x[m])
			den := suite.Scalar().Sub(xjScalar, xmScalar)
			denInv := suite.Scalar().Inv(den)
			num := suite.Scalar().Sub(suite.Scalar().Zero(), xmScalar)
			term := suite.Scalar().Mul(num, denInv)
			basis = suite.Scalar().Mul(basis, term)
		}

		// 乘以对应的y值并累加到结果中
		term := suite.Scalar().Mul(basis, y[j])
		for i := range resultCoeffs {
			resultCoeffs[i] = suite.Scalar().Add(resultCoeffs[i], term)
		}
	}

	return &PolyKyberImpl{coeff: resultCoeffs}, nil
}
