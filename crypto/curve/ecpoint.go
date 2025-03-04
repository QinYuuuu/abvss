<<<<<<< HEAD
package curve

import (
	"math/big"
)

type ECPoint struct {
	x *big.Int
	y *big.Int
}

func NewECPoint(x, y *big.Int) *ECPoint {
	return &ECPoint{
		x: x,
		y: y,
	}
}

func (p ECPoint) X() *big.Int {
	return p.x
}

func (p ECPoint) Y() *big.Int {
	return p.y
}
=======
package curve

import (
	"math/big"
)

type ECPoint struct {
	x *big.Int
	y *big.Int
}

func NewECPoint(x, y *big.Int) *ECPoint {
	return &ECPoint{
		x: x,
		y: y,
	}
}

func (p ECPoint) X() *big.Int {
	return p.x
}

func (p ECPoint) Y() *big.Int {
	return p.y
}
>>>>>>> 19b0d27 (Initial commit)
