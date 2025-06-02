package pkg

import (
	"fmt"

	"go.dedis.ch/kyber/v4"
)

// InterpolationKyber 在KyberPoint下重新实现Lagrange插值函数
func InterpolationKyber(x []int64, y []kyber.Scalar, group kyber.Group) (*PolyKyberImpl, error) {
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
		resultCoeffs[i] = group.Scalar().Zero()
	}

	for j := 0; j < len(x); j++ {
		basis := group.Scalar().One()
		for m := 0; m < len(x); m++ {
			if m == j {
				continue
			}
			// 原代码中 x[j] 和 x[m] 是 kyber.Scalar 类型，现在 x 是 []int64，需要转换
			xjScalar := group.Scalar().SetInt64(x[j])
			xmScalar := group.Scalar().SetInt64(x[m])
			den := group.Scalar().Sub(xjScalar, xmScalar)
			denInv := group.Scalar().Inv(den)
			num := group.Scalar().Sub(group.Scalar().Zero(), xmScalar)
			term := group.Scalar().Mul(num, denInv)
			basis = group.Scalar().Mul(basis, term)
		}

		// 乘以对应的y值并累加到结果中
		term := group.Scalar().Mul(basis, y[j])
		for i := range resultCoeffs {
			resultCoeffs[i] = group.Scalar().Add(resultCoeffs[i], term)
		}
	}

	return &PolyKyberImpl{group: group, coeff: resultCoeffs}, nil
}
