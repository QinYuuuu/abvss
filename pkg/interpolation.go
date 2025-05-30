package pkg

import (
	"fmt"

	"go.dedis.ch/kyber/v4"
)

// InterpolationAtZeroKyberPoint
// input: (x_i, g^(y_i))
// return: g^(y_0), error
func InterpolationAtZeroKyberPoint(x []int64, y []kyber.Point, group kyber.Group) (kyber.Point, error) {
	if len(x) != len(y) {
		return nil, fmt.Errorf("x and y not same length")
	}
	n := len(x)
	if len(x) == 0 {
		return nil, fmt.Errorf("x and y must not be empty")
	}
	for i := 0; i < n; i++ {
		if x[i] == 0 {
			return y[i], nil
		}
	}
	if len(x) == 1 {
		return nil, fmt.Errorf("at least two x value to interpolate")
	}
	estimatedXAtYZero := group.Point().Mul(group.Scalar().Zero(), nil)
	for j := 0; j < n; j++ {
		xj := group.Scalar().SetInt64(x[j])
		yj := y[j]
		ljAtZero := group.Scalar().One()
		for i := 0; i < n; i++ {
			if i == j {
				continue
			}
			xi := group.Scalar().SetInt64(x[i])

			// L_j(0) = Π (k≠j) [ x_k / (x_k - x_j) ]
			denominator := group.Scalar().Sub(xi, xj)
			if x[i] == x[j] {
				return nil, fmt.Errorf("calculate L_%d(0): x_%d (%f) - x_%d (%f) == 0", j, i, xi, j, xj)
			}
			ljAtZero = group.Scalar().Mul(ljAtZero, group.Scalar().Div(xi, denominator))
		}
		tmp := group.Point().Mul(ljAtZero, yj)
		estimatedXAtYZero = group.Point().Add(estimatedXAtYZero, tmp)
	}
	return estimatedXAtYZero, nil
}
