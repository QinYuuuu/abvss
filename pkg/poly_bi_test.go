package pkg

import (
	"math/big"
	"testing"

	"gotest.tools/v3/assert"
)

func TestBiPolyEvalAtXMod(t *testing.T) {
	p := big.NewInt(17)
	poly := &BivariatePoly{
		coeff: [][]*big.Int{
			{big.NewInt(3), big.NewInt(1)}, // 3*y^0 + 1*y^1
			{big.NewInt(2), big.NewInt(4)}, // 2*x^1*y^0 +4*x^1*y^1
		},
		degreeX: 1,
		degreeY: 1,
	}

	// 计算在x=2处的y多项式
	yPoly := poly.EvalAtXMod(big.NewInt(2), p)
	want := &PolyBigIntImpl{
		coeff: []*big.Int{big.NewInt(7), big.NewInt(9)},
	}
	assert.Equal(t, want.Equal(yPoly), true)
}

func TestNewRandBiPoly(t *testing.T) {
	degreeX := 2
	degreeY := 3
	n := big.NewInt(100)
	poly, err := NewRandBiPoly(degreeX, degreeY, n)
	if err != nil {
		t.Errorf("NewRandBiPoly failed: %v", err)
	}
	if poly.GetDegreeX() != degreeX || poly.GetDegreeY() != degreeY {
		t.Errorf("Degree mismatch, expected X: %d, Y: %d, got X: %d, Y: %d", degreeX, degreeY, poly.GetDegreeX(), poly.GetDegreeY())
	}
}

func TestNewConstantBiPoly(t *testing.T) {
	c := int64(5)
	poly := NewConstantBiPoly(c)
	coeff, err := poly.GetCoefficient(0, 0)
	if err != nil {
		t.Errorf("GetCoefficient failed: %v", err)
	}
	if coeff.Int64() != c {
		t.Errorf("Coefficient mismatch, expected: %d, got: %d", c, coeff.Int64())
	}
}

func TestGetDegreeX(t *testing.T) {
	poly, _ := NewBiPoly(2, 2)
	poly.SetCoefficient(1, 1, 1)
	if poly.GetDegreeX() != 1 {
		t.Errorf("GetDegreeX failed, expected: 1, got: %d", poly.GetDegreeX())
	}
}

func TestGetDegreeY(t *testing.T) {
	poly, _ := NewBiPoly(2, 2)
	poly.SetCoefficient(1, 1, 1)
	if poly.GetDegreeY() != 1 {
		t.Errorf("GetDegreeY failed, expected: 1, got: %d", poly.GetDegreeY())
	}
}

func TestGetCoefficient(t *testing.T) {
	poly, _ := NewBiPoly(1, 1)
	poly.SetCoefficient(0, 0, 1)
	coeff, err := poly.GetCoefficient(0, 0)
	if err != nil {
		t.Errorf("GetCoefficient failed: %v", err)
	}
	if coeff.Int64() != 1 {
		t.Errorf("Coefficient mismatch, expected: 1, got: %d", coeff.Int64())
	}
}

func TestSetCoefficient(t *testing.T) {
	poly, _ := NewBiPoly(1, 1)
	err := poly.SetCoefficient(0, 0, 1)
	if err != nil {
		t.Errorf("SetCoefficient failed: %v", err)
	}
	coeff, _ := poly.GetCoefficient(0, 0)
	if coeff.Int64() != 1 {
		t.Errorf("Coefficient mismatch, expected: 1, got: %d", coeff.Int64())
	}
}

func TestSetCoefficientBig(t *testing.T) {
	poly, _ := NewBiPoly(1, 1)
	c := big.NewInt(1)
	err := poly.SetCoefficientBig(0, 0, c)
	if err != nil {
		t.Errorf("SetCoefficientBig failed: %v", err)
	}
	coeff, _ := poly.GetCoefficient(0, 0)
	if coeff.Cmp(c) != 0 {
		t.Errorf("Coefficient mismatch, expected: %d, got: %d", c.Int64(), coeff.Int64())
	}
}

func TestReset(t *testing.T) {
	poly, _ := NewBiPoly(1, 1)
	poly.SetCoefficient(0, 0, 1)
	poly.Reset()
	if !poly.IsZero() {
		t.Errorf("Reset failed, polynomial should be zero")
	}
}

func TestDeepCopy(t *testing.T) {
	poly1, _ := NewBiPoly(1, 1)
	poly1.SetCoefficient(0, 0, 1)
	poly2, _ := NewBiPoly(0, 0)
	poly2.DeepCopy(poly1)
	if !poly2.Equal(poly1) {
		t.Errorf("DeepCopy failed, polynomials should be equal")
	}
}

func TestEqual(t *testing.T) {
	poly1, _ := NewBiPoly(1, 1)
	poly2, _ := NewBiPoly(1, 1)
	poly1.SetCoefficient(0, 0, 1)
	poly2.SetCoefficient(0, 0, 1)
	if !poly1.Equal(poly2) {
		t.Errorf("Equal failed, polynomials should be equal")
	}
}

func TestIsZero(t *testing.T) {
	poly, _ := NewBiPoly(1, 1)
	if !poly.IsZero() {
		t.Errorf("IsZero failed, polynomial should be zero")
	}
}
