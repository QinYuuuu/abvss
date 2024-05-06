package curve

import "math/big"

type CurveParams interface {
}

type Curve interface {
	// Params返回曲线的参数
	Params() *CurveParams
	// IsOnCurve verify if (x，y) on the curve
	IsOnCurve(x, y *big.Int) bool
	// Add返回(x1,y1)和(x2,y2)的和
	Add(x1, y1, x2, y2 *big.Int) (x, y *big.Int)
	// Double返回 2*(x,y)
	Double(x1, y1 *big.Int) (x, y *big.Int)
	// ScalarMult返回k*(Bx,By) and k in big-endian
	ScalarMult(x1, y1 *big.Int, k []byte) (x, y *big.Int)
	// ScalarBaseMult返回 k*G, G是组的基点。
	// k是大端形式的整数。
	ScalarBaseMult(k []byte) (x, y *big.Int)
}
