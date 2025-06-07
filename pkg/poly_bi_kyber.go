package pkg

import (
	"errors"
	"fmt"

	"go.dedis.ch/kyber/v4/util/random"

	"go.dedis.ch/kyber/v4"
)

// Assume the existence of the BivariatePolyKyberImpl structure
type BivariatePolyKyberImpl struct {
	group   kyber.Group
	coeff   [][]kyber.Scalar
	degreeX int
	degreeY int
}

// NewBiPolyKyber returns a bivariate polynomial P(x,y) = 0 with specified degrees of x and y
func NewBiPolyKyber(degreeX, degreeY int, group kyber.Group) (*BivariatePolyKyberImpl, error) {
	if degreeX < 0 || degreeY < 0 {
		return nil, fmt.Errorf("degreeX and degreeY must be non-negative, got degreeX: %d, degreeY: %d", degreeX, degreeY)
	}

	coeff := make([][]kyber.Scalar, degreeX+1)
	for i := range coeff {
		coeff[i] = make([]kyber.Scalar, degreeY+1)
		for j := range coeff[i] {
			zero := group.Scalar().Zero()
			coeff[i][j] = zero
		}
	}

	return &BivariatePolyKyberImpl{
		group:   group,
		coeff:   coeff,
		degreeX: degreeX,
		degreeY: degreeY,
	}, nil
}

// NewRandBiPolyKyber returns a randomized bivariate polynomial with specified degrees of x and y
// Coefficients are pseudo-random numbers in [0, n) (这里n的概念可能需要调整以适应kyber)
func NewRandBiPolyKyber(degreeX, degreeY int, group kyber.Group) (*BivariatePolyKyberImpl, error) {
	p, e := NewBiPolyKyber(degreeX, degreeY, group)
	if e != nil {
		return nil, e
	}

	for i := range p.coeff {
		for j := range p.coeff[i] {
			p.coeff[i][j] = group.Scalar().Pick(random.New())
		}
	}

	return p, nil
}

// NewConstantBiPolyKyber creates a constant bivariate polynomial P(x,y) = c
func NewConstantBiPolyKyber(suite kyber.Group, c int64) *BivariatePolyKyberImpl {
	zero, err := NewBiPolyKyber(0, 0, suite)
	if err != nil {
		panic(err.Error())
	}

	constant := suite.Scalar().SetInt64(c)
	zero.coeff[0][0] = constant
	return zero
}

// GetDegreeX returns the degree of x, ignoring leading zeros
func (poly *BivariatePolyKyberImpl) GetDegreeX() int {
	deg := poly.degreeX
	for i := deg; i > 0; i-- {
		nonZero := false
		for j := 0; j <= poly.degreeY; j++ {
			if !poly.coeff[i][j].Equal(poly.coeff[i][j].Zero()) {
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
func (poly *BivariatePolyKyberImpl) GetDegreeY() int {
	deg := poly.degreeY
	for i := deg; i > 0; i-- {
		nonZero := false
		for j := 0; j <= poly.degreeX; j++ {
			if !poly.coeff[j][i].Equal(poly.coeff[j][i].Zero()) {
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
func (poly *BivariatePolyKyberImpl) GetCoefficient(i, j int) (kyber.Scalar, error) {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		zero := poly.coeff[0][0].Zero()
		return zero, errors.New("out of range")
	}

	return poly.coeff[i][j], nil
}

// SetCoefficient sets poly.coeff[i][j] to ci
func (poly *BivariatePolyKyberImpl) SetCoefficient(i, j int, suite kyber.Group, ci int64) error {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		return errors.New("out of border")
	}

	newScalar := suite.Scalar().SetInt64(ci)
	poly.coeff[i][j] = newScalar
	return nil
}

// SetCoefficientBig sets poly.coeff[i][j] to ci (a big.Int) 这里可能需要调整为kyber.Scalar
func (poly *BivariatePolyKyberImpl) SetCoefficientScalar(i, j int, ci kyber.Scalar) error {
	if i < 0 || i > poly.degreeX || j < 0 || j > poly.degreeY {
		return errors.New("out of border")
	}

	poly.coeff[i][j] = ci
	return nil
}

// Reset sets the coefficients to zero
func (poly *BivariatePolyKyberImpl) Reset(suite kyber.Group) {
	for i := range poly.coeff {
		for j := range poly.coeff[i] {
			zero := suite.Scalar().Zero()
			poly.coeff[i][j] = zero
		}
	}
}

func (poly *BivariatePolyKyberImpl) DeepCopy(suite kyber.Group, other *BivariatePolyKyberImpl) {
	poly.resetToDegree(suite, other.degreeX, other.degreeY)

	for i := 0; i <= other.degreeX; i++ {
		for j := 0; j <= other.degreeY; j++ {
			newScalar := suite.Scalar().Set(other.coeff[i][j])
			poly.coeff[i][j] = newScalar
		}
	}
}

// resetToDegree resizes the coefficient matrix to the specified degrees
func (poly *BivariatePolyKyberImpl) resetToDegree(suite kyber.Group, degreeX, degreeY int) {
	if degreeX+1 <= len(poly.coeff) {
		poly.coeff = poly.coeff[:degreeX+1]
	} else {
		neededX := degreeX + 1 - len(poly.coeff)
		neededRows := make([][]kyber.Scalar, neededX)
		for i := range neededRows {
			neededRows[i] = make([]kyber.Scalar, degreeY+1)
			for j := range neededRows[i] {
				zero := suite.Scalar().Zero()
				neededRows[i][j] = zero
			}
		}
		poly.coeff = append(poly.coeff, neededRows...)
	}

	for i := range poly.coeff {
		if degreeY+1 <= len(poly.coeff[i]) {
			poly.coeff[i] = poly.coeff[i][:degreeY+1]
		} else {
			neededY := degreeY + 1 - len(poly.coeff[i])
			neededCols := make([]kyber.Scalar, neededY)
			for j := range neededCols {
				zero := suite.Scalar().Zero()
				neededCols[j] = zero
			}
			poly.coeff[i] = append(poly.coeff[i], neededCols...)
		}
	}

	poly.Reset(suite)
}

func (poly *BivariatePolyKyberImpl) Equal(suite kyber.Group, op *BivariatePolyKyberImpl) bool {
	if op.GetDegreeX() != poly.GetDegreeX() || op.GetDegreeY() != poly.GetDegreeY() {
		return false
	}

	for i := 0; i <= op.GetDegreeX(); i++ {
		for j := 0; j <= op.GetDegreeY(); j++ {
			if !poly.coeff[i][j].Equal(op.coeff[i][j]) {
				return false
			}
		}
	}

	return true
}

// IsZero returns whether poly is a zero polynomial
func (poly *BivariatePolyKyberImpl) IsZero(suite kyber.Group) bool {
	zero := suite.Scalar().Zero()
	for i := 0; i <= poly.degreeX; i++ {
		for j := 0; j <= poly.degreeY; j++ {
			if !poly.coeff[i][j].Equal(zero) {
				return false
			}
		}
	}

	return true
}

// EvalAtYMod evaluates the polynomial at a given y value modulo a provided modulus.
func (poly *BivariatePolyKyberImpl) EvalAtYMod(y kyber.Scalar) *PolyKyberImpl {
	result := NewEmptyKyber(poly.group)
	result.resetToDegree(poly.group, poly.degreeX)

	for i := 0; i <= poly.degreeX; i++ {
		term := poly.group.Scalar().Zero()
		for j := 0; j <= poly.degreeY; j++ {
			tmp := poly.group.Scalar().SetInt64(int64(j))
			yTerm := poly.group.Scalar().Set(poly.coeff[i][j])
			tmp.Mul(tmp, yTerm)
			term.Add(term, tmp)
		}
		result.SetCoefficientScalar(i, term)
	}
	return result
}

// EvalAtXMod 在给定x值和模数下计算多项式的值
func (poly *BivariatePolyKyberImpl) EvalAtXMod(x kyber.Scalar) *PolyKyberImpl {
	result := NewEmptyKyber(poly.group)
	result.resetToDegree(poly.group, poly.degreeY)
	for i := 0; i <= poly.degreeY; i++ {
		term := poly.group.Scalar().Zero()
		xExpi := poly.group.Scalar().One()
		for j := 0; j <= poly.degreeX; j++ {
			xCoeff := poly.group.Scalar().Set(poly.coeff[j][i])
			tmp := poly.group.Scalar().Mul(xExpi, xCoeff)
			xExpi = poly.group.Scalar().Mul(xExpi, x)
			term.Add(term, tmp)
		}
		result.SetCoefficientScalar(i, term)
	}
	return result
}

func (poly *BivariatePolyKyberImpl) ToString() string {
	s := ""
	for i := poly.degreeX; i >= 0; i-- {
		for j := poly.degreeY; j >= 0; j-- {
			if i > 0 || j > 0 {
				s += fmt.Sprintf("%s x^%dy^%d + ", poly.coeff[i][j].String(), i, j)
			}
			if i == 0 && j == 0 {
				s += poly.coeff[i][j].String()
			}
		}
	}
	return s
}
