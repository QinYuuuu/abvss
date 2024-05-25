package curve

import (
	"crypto/elliptic"
	"errors"
	"math/big"
)

var (
	one  = big.NewInt(1)
	zero = big.NewInt(0)
)

type Curve interface {
	// Params 返回曲线的参数
	Params() *elliptic.CurveParams
	// IsOnCurve verify if (x，y) on the curve
	IsOnCurve(x, y *big.Int) bool
	// Add 返回(x1,y1)和(x2,y2)的和
	Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int)
	// Double 返回 2*(x,y)
	Double(x1, y1 *big.Int) (x, y *big.Int)
	// ScalarMult 返回k*(Bx,By) and k in big-endian
	ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int)
	// ScalarBaseMult 返回 k*G, G是组的基点。
	// k是大端形式的整数。
	ScalarBaseMult(k []byte) (x, y *big.Int)
}

func DotProductGroup(curve Curve, v1 []*big.Int, v2x, v2y []*big.Int) (*big.Int, *big.Int, error) {
	if len(v1) != len(v2x) || len(v1) != len(v2y) || len(v2x) != len(v2y) {
		return nil, nil, errors.New("the input length is different")
	}
	dotx, doty := curve.ScalarMult(v2x[0], v2y[0], v1[0].Bytes())
	for i := 1; i < len(v1); i++ {
		tmpx, tmpy := curve.ScalarMult(v2x[i], v2y[i], v1[i].Bytes())
		dotx, doty = curve.Add(dotx, doty, tmpx, tmpy)
	}
	return dotx, doty, nil
}
