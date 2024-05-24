package polynomial

import (
	"bytes"
	"fmt"
	"math/big"
)

func arrayOfZeroes(n int) []*big.Int {
	r := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		r[i] = new(big.Int).SetInt64(0)
	}
	return r[:]
}

func fAdd(a, b *big.Int, R *big.Int) *big.Int {
	ab := new(big.Int).Add(a, b)
	return ab.Mod(ab, R)
}

func fSub(a, b *big.Int, R *big.Int) *big.Int {
	ab := new(big.Int).Sub(a, b)
	return new(big.Int).Mod(ab, R)
}

func fMul(a, b *big.Int, R *big.Int) *big.Int {
	ab := new(big.Int).Mul(a, b)
	return ab.Mod(ab, R)
}

func fDiv(a, b, R *big.Int) *big.Int {
	ab := new(big.Int).Mul(a, new(big.Int).ModInverse(b, R))
	return new(big.Int).Mod(ab, R)
}

func fNeg(a *big.Int, R *big.Int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Neg(a), R)
}

func fExp(base *big.Int, e *big.Int, R *big.Int) *big.Int {
	res := big.NewInt(1)
	rem := new(big.Int).Set(e)
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		// if BigIsOdd(rem) {
		if rem.Bit(0) == 1 { // .Bit(0) returns 1 when is odd
			res = fMul(res, exp, R)
		}
		exp = fMul(exp, exp, R)
		rem.Rsh(rem, 1)
	}
	return res
}

func polynomialAdd(a, b []*big.Int, R *big.Int) []*big.Int {
	r := arrayOfZeroes(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = fAdd(r[i], a[i], R)
	}
	for i := 0; i < len(b); i++ {
		r[i] = fAdd(r[i], b[i], R)
	}
	return r
}

func polynomialSub(a, b []*big.Int, R *big.Int) []*big.Int {
	r := arrayOfZeroes(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = fAdd(r[i], a[i], R)
	}
	for i := 0; i < len(b); i++ {
		r[i] = fSub(r[i], b[i], R)
	}
	return r
}

func polynomialMul(a, b []*big.Int, R *big.Int) []*big.Int {
	r := arrayOfZeroes(len(a) + len(b) - 1)
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			r[i+j] = fAdd(r[i+j], fMul(a[i], b[j], R), R)
		}
	}
	return r
}

func polynomialDiv(a, b []*big.Int, R *big.Int) ([]*big.Int, []*big.Int) {
	// https://en.wikipedia.org/wiki/Division_algorithm
	r := arrayOfZeroes(len(a) - len(b) + 1)
	rem := a
	for len(rem) >= len(b) {
		l := fDiv(rem[len(rem)-1], b[len(b)-1], R)
		pos := len(rem) - len(b)
		r[pos] = l
		aux := arrayOfZeroes(pos)
		aux1 := append(aux, l)
		aux2 := polynomialSub(rem, polynomialMul(b, aux1, R), R)
		rem = aux2[:len(aux2)-1]
	}
	return r, rem
}

func polynomialMulByConstant(a []*big.Int, c, R *big.Int) []*big.Int {
	for i := 0; i < len(a); i++ {
		a[i] = fMul(a[i], c, R)
	}
	return a
}
func polynomialDivByConstant(a []*big.Int, c, R *big.Int) []*big.Int {
	for i := 0; i < len(a); i++ {
		a[i] = fDiv(a[i], c, R)
	}
	return a
}

// LagrangeInterpolation implements the Lagrange interpolation:
// https://en.wikipedia.org/wiki/Lagrange_polynomial
func LagrangeInterpolation(x, y []*big.Int, R *big.Int) (Polynomial, error) {
	// p(x) will be the interpoled polynomial
	// var p []*big.Int
	if len(x) != len(y) {
		return Polynomial{}, fmt.Errorf("len(x)!=len(y): %d, %d", len(x), len(y))
	}
	p := arrayOfZeroes(len(x))
	k := len(x)

	for j := 0; j < k; j++ {
		// jPol is the Lagrange basis polynomial for each point
		var jPol []*big.Int
		for m := 0; m < k; m++ {
			// if x[m] == x[j] {
			if m == j {
				continue
			}
			// numerator & denominator of the current iteration
			num := []*big.Int{fNeg(x[m], R), big.NewInt(1)} // (x^1 - x_m)
			den := fSub(x[j], x[m], R)                      // x_j-x_m
			mPol := polynomialDivByConstant(num, den, R)
			if len(jPol) == 0 {
				// first j iteration
				jPol = mPol
				continue
			}
			jPol = polynomialMul(jPol, mPol, R)
		}
		p = polynomialAdd(p, polynomialMulByConstant(jPol, y[j], R), R)
	}
	result := Polynomial{coeff: p}
	return result, nil
}
