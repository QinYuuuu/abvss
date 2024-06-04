package polynomial

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ZERO := big.NewInt(0)

	degree := 100
	poly, err := New(degree)

	assert.Nil(t, err, "error in New")
	assert.Equal(t, degree+1, len(poly.coeff), "coeff len")

	for i := 0; i < len(poly.coeff); i++ {
		assert.Zero(t, poly.coeff[i].Cmp(ZERO))
	}

	_, err = New(-1)
	assert.NotNil(t, err, "negative degree")
}

func TestNewOne(t *testing.T) {
	ONE := big.NewInt(1)

	onePoly := NewOne()

	assert.Equal(t, 0, onePoly.GetDegree(), "degree")
	assert.Equal(t, 1, len(onePoly.coeff), "coeff len")

	assert.Equal(t, 0, ONE.Cmp(onePoly.coeff[0]))
}

func TestNewEmpty(t *testing.T) {
	emptyPoly := NewEmpty()

	assert.Equal(t, 0, emptyPoly.GetDegree(), "degree")
	assert.Equal(t, int64(0), emptyPoly.coeff[0].Int64(), "const")
}

func TestNewRand(t *testing.T) {
	var degree = 100
	var n = big.NewInt(1000)

	poly, err := NewRand(degree, n)
	assert.Nil(t, err, "err in NewRand")

	assert.Equal(t, degree+1, len(poly.coeff), "coeff len")

	for i := range poly.coeff {
		assert.Equal(t, -1, poly.coeff[i].Cmp(n), "rand range")
	}
}

func TestPolynomialCap(t *testing.T) {
	op := FromVec(1, 1, 1, 1, 1, 1, 0, 0, 0)
	assert.Equal(t, 9, op.GetCap())

	op.GrowCapTo(100)
	assert.Equal(t, 100, op.GetCap())
}

func TestPolynomial_Add(t *testing.T) {
	var degree = 10
	var n = big.NewInt(1000)

	poly1, err := NewRand(degree, n)
	assert.Nil(t, err, "err in NewRand")

	poly2, err := NewRand(degree, n)
	assert.Nil(t, err, "err in NewRand")

	result := NewEmpty()

	err = result.Add(poly1, poly2)
	assert.Nil(t, err, "add")

	var tmp = big.NewInt(0)
	for i := 0; i <= degree; i++ {
		tmp.Add(poly1.coeff[i], poly2.coeff[i])
		assert.Zero(t, result.coeff[i].Cmp(tmp), "add result")
		tmp.SetInt64(0)
	}
}

func TestPolynomial_Sub(t *testing.T) {
	var tests = []struct {
		op1      []int64
		op2      []int64
		expected []int64
	}{
		{[]int64{1, 1}, []int64{0, 1}, []int64{1}},
		{[]int64{1, 1, 1}, []int64{1, 1, 1}, []int64{0}},
	}

	for _, test := range tests {
		op1 := FromVec(test.op1...)
		op2 := FromVec(test.op2...)
		expected := FromVec(test.expected...)

		result, _ := New(op1.GetDegree())
		result.Sub(op1, op2)

		assert.True(t, expected.Equal(result))
	}

}

func TestPolynomial_Mul(t *testing.T) {
	op1 := FromVec(1, 1, 1, 1, 1, 1)
	result := NewEmpty()

	err := result.Mul(op1, op1)
	assert.Nil(t, err, "Mul")

	expected := FromVec(1, 2, 3, 4, 5, 6, 5, 4, 3, 2, 1)
	assert.True(t, expected.Equal(result), "Mul")
}

func TestDiv(t *testing.T) {
	mod := big.NewInt(17)

	// to test if q, r = DivMod(a, b)
	var tests = []struct {
		a []int64
		b []int64
		q []int64
		r []int64
	}{
		{[]int64{1, 2, 1}, []int64{1, 1}, []int64{1, 1}, []int64{}},
		{[]int64{7, 0, 0, 0, 2, 1}, []int64{-5, 0, 0, 1}, []int64{0, 2, 1}, []int64{7, 10, 5}},
		{[]int64{7, 10, 5, 2}, []int64{4, 0, 1}, []int64{5, 2}, []int64{4, 2}},
		{[]int64{1, 2, 1}, []int64{1, 2}, []int64{5, 9}, []int64{13}},
	}

	for _, test := range tests {
		a := FromVec(test.a...)
		b := FromVec(test.b...)
		q := FromVec(test.q...)
		r := FromVec(test.r...)

		qq, rr, err := DivMod(a, b, mod)
		assert.Nil(t, err, "DivMod")

		assert.True(t, qq.Equal(q))
		assert.True(t, rr.Equal(r))
	}
}

func TestPolynomial_EvalMod(t *testing.T) {
	var tests = []struct {
		coeffs   []int64
		evalAt   int64
		expected int64
	}{
		{[]int64{1, 1}, 1, 2},
		{[]int64{1, 2, 3}, 0, 1},
		{[]int64{1, 2, 3}, 1, 6},
		{[]int64{1, 2, 3}, 2, 17},
		{[]int64{2, 2}, 0, 2},
		{[]int64{1, 0, 0, 1, 0}, 3, 28},
	}

	mod := big.NewInt(100)

	for _, test := range tests {
		p := FromVec(test.coeffs...)
		eval := p.EvalMod(big.NewInt(test.evalAt), mod)
		assert.Equal(t, test.expected, eval.Int64(), p.ToString())
	}
}
