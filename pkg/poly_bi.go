package pkg

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

// Assume the existence of the BivariatePoly structure
type BivariatePoly struct {
	coeff   [][]*big.Int
	degreeX int
	degreeY int
}

// NewBiPoly returns a bivariate polynomial P(x,y) = 0 with specified degrees of x and y
func NewBiPoly(degreeX, degreeY int) (*BivariatePoly, error) {
	if degreeX < 0 || degreeY < 0 {
		return nil, fmt.Errorf("degreeX and degreeY must be non-negative, got degreeX: %d, degreeY: %d", degreeX, degreeY)
	}

	coeff := make([][]*big.Int, degreeX+1)
	for i := range coeff {
		coeff[i] = make([]*big.Int, degreeY+1)
		for j := range coeff[i] {
			coeff[i][j] = big.NewInt(0)
		}
	}

	return &BivariatePoly{coeff: coeff, degreeX: degreeX, degreeY: degreeY}, nil
}

// NewRandBiPoly returns a randomized bivariate polynomial with specified degrees of x and y
// Coefficients are pseudo-random numbers in [0, n)
func NewRandBiPoly(degreeX, degreeY int, n *big.Int) (*BivariatePoly, error) {
	p, e := NewBiPoly(degreeX, degreeY)
	if e != nil {
		return nil, e
	}

	p.Rand(n)

	return p, nil
}

// NewConstantBiPoly creates a constant bivariate polynomial P(x,y) = c
func NewConstantBiPoly(c int64) *BivariatePoly {
	zero, err := NewBiPoly(0, 0)
	if err != nil {
		panic(err.Error())
	}

	zero.coeff[0][0] = big.NewInt(c)
	return zero
}

// GetDegreeX returns the degree of x, ignoring leading zeros
func (poly *BivariatePoly) GetDegreeX() int {
	deg := poly.degreeX
	for i := deg; i > 0; i-- {
		nonZero := false
		for j := 0; j <= poly.degreeY; j++ {
			if poly.coeff[i][j].Int64() != 0 {
				nonZero = true
				break
			}
		}
		if nonZero {
			break
		}
		deg--
	}
	return deg
}

// GetDegreeY returns the degree of y, ignoring leading zeros
func (poly *BivariatePoly) GetDegreeY() int {
	deg := poly.degreeY
	for i := deg; i > 0; i-- {
		nonZero := false
		for j := 0; j <= poly.degreeX; j++ {
			if poly.coeff[j][i].Int64() != 0 {
				nonZero = true
				break
			}
		}
		if nonZero {
			break
		}
		deg--
	}
	return deg
}

// GetCoefficient returns coeff[i][j]
func (poly *BivariatePoly) GetCoefficient(i, j int) (*big.Int, error) {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		return big.NewInt(0), errors.New("超出边界")
	}

	return poly.coeff[i][j], nil
}

// SetCoefficient sets poly.coeff[i][j] to ci
func (poly *BivariatePoly) SetCoefficient(i, j int, ci int64) error {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		return errors.New("超出边界")
	}

	poly.coeff[i][j].SetInt64(ci)
	return nil
}

// SetCoefficientBig sets poly.coeff[i][j] to ci (a big.Int)
func (poly *BivariatePoly) SetCoefficientBig(i, j int, ci *big.Int) error {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		return errors.New("超出边界")
	}

	poly.coeff[i][j].Set(ci)
	return nil
}

// Reset sets the coefficients to zero
func (poly *BivariatePoly) Reset() {
	for i := range poly.coeff {
		for j := range poly.coeff[i] {
			poly.coeff[i][j].SetInt64(0)
		}
	}
}

func (poly *BivariatePoly) DeepCopy(other *BivariatePoly) {
	poly.resetToDegree(other.degreeX, other.degreeY)

	for i := 0; i <= other.degreeX; i++ {
		for j := 0; j <= other.degreeY; j++ {
			poly.coeff[i][j].Set(other.coeff[i][j])
		}
	}
}

// resetToDegree resizes the coefficient matrix to the specified degrees
func (poly *BivariatePoly) resetToDegree(degreeX, degreeY int) {
	if degreeX+1 <= len(poly.coeff) {
		poly.coeff = poly.coeff[:degreeX+1]
	} else {
		neededX := degreeX + 1 - len(poly.coeff)
		neededRows := make([][]*big.Int, neededX)
		for i := range neededRows {
			neededRows[i] = make([]*big.Int, degreeY+1)
			for j := range neededRows[i] {
				neededRows[i][j] = big.NewInt(0)
			}
		}
		poly.coeff = append(poly.coeff, neededRows...)
	}

	for i := range poly.coeff {
		if degreeY+1 <= len(poly.coeff[i]) {
			poly.coeff[i] = poly.coeff[i][:degreeY+1]
		} else {
			neededY := degreeY + 1 - len(poly.coeff[i])
			neededCols := make([]*big.Int, neededY)
			for j := range neededCols {
				neededCols[j] = big.NewInt(0)
			}
			poly.coeff[i] = append(poly.coeff[i], neededCols...)
		}
	}

	poly.Reset()
}

func (poly *BivariatePoly) Equal(op *BivariatePoly) bool {
	if op.GetDegreeX() != poly.GetDegreeX() || op.GetDegreeY() != poly.GetDegreeY() {
		return false
	}

	for i := 0; i <= op.GetDegreeX(); i++ {
		for j := 0; j <= op.GetDegreeY(); j++ {
			if op.coeff[i][j].Cmp(poly.coeff[i][j]) != 0 {
				return false
			}
		}
	}

	return true
}

// IsZero returns whether poly is a zero polynomial
func (poly *BivariatePoly) IsZero() bool {
	for i := 0; i <= poly.degreeX; i++ {
		for j := 0; j <= poly.degreeY; j++ {
			if poly.coeff[i][j].Int64() != 0 {
				return false
			}
		}
	}
	return true
}

// Rand 将多项式的系数设置为 [0, n) 内的伪随机数
func (poly *BivariatePoly) Rand(mod *big.Int) {
	for i := range poly.coeff {
		for j := range poly.coeff[i] {
			poly.coeff[i][j], _ = rand.Int(rand.Reader, mod)
		}
	}
}

// EvalAtXMod evaluates the polynomial at a given x value and takes the result modulo p.
func (poly *BivariatePoly) EvalAtXMod(x *big.Int, p *big.Int) *PolyBigIntImpl {
	degreeY := poly.GetDegreeY()
	result := NewEmpty()
	result.GrowCapTo(degreeY + 1)

	for j := 0; j <= degreeY; j++ {
		term := big.NewInt(0)
		for i := 0; i <= poly.GetDegreeX(); i++ {
			tmp := new(big.Int).Exp(x, big.NewInt(int64(i)), nil)
			tmp.Mul(tmp, poly.coeff[i][j])
			term.Add(term, tmp)
		}
		if p != nil {
			term.Mod(term, p)
		}
		result.SetCoefficientBig(j, term)
	}

	return result
}

// EvalAtYMod evaluates the polynomial at a given y value and takes the result modulo p.
func (poly *BivariatePoly) EvalAtYMod(y *big.Int, p *big.Int) *PolyBigIntImpl {
	degreeX := poly.GetDegreeX()
	result := NewEmpty()
	result.GrowCapTo(degreeX + 1)

	for i := 0; i <= degreeX; i++ {
		term := big.NewInt(0)
		for j := 0; j <= poly.GetDegreeY(); j++ {
			tmp := new(big.Int).Exp(y, big.NewInt(int64(j)), nil)
			tmp.Mul(tmp, poly.coeff[i][j])
			term.Add(term, tmp)
		}
		if p != nil {
			term.Mod(term, p)
		}
		result.SetCoefficientBig(i, term)
	}

	return result
}
