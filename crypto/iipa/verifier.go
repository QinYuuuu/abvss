package iipa

import "go.dedis.ch/kyber/v3"

type Verifier struct {
	crs    *CRS
	input1 chan []byte
	input2 chan kyber.Point
	kyber.Group
}

func NewVerifier() *Verifier {
	return &Verifier{
		input1: make(chan []byte),
		input2: make(chan kyber.Point),
	}
}

func (v *Verifier) RecursiveVerify(gVec []kyber.Point, h, P kyber.Point, n int) bool {
	// step 1
	// 1.1 Verifier receive aVec from Prover (length of aVec == 1)
	// 1.2 verify P
	if n == 1 {
		var a, b kyber.Scalar
		// get from input channel
		a = v.Scalar().SetBytes(<-v.input1)
		b = v.Scalar().SetBytes(<-v.input1)

		left := v.Point().Mul(a, gVec[0])
		right := v.Point().Mul(v.Scalar().Mul(a, b), h)
		PWant := v.Point().Add(left, right)
		return P.Equal(PWant)
	}

	// step 2
	// 2.1 Verifier receive aVec[n], bVec[n] from Prover
	// 2.3 update P, aVec, gVec, n
	if n%2 == 1 {
		var aVecN, bVecN kyber.Scalar
		// get from input channel
		aVecN = v.Scalar().SetBytes(<-v.input1)
		bVecN = v.Scalar().SetBytes(<-v.input1)

		aNeg := v.Scalar().Neg(aVecN)
		bNeg := v.Scalar().Neg(bVecN)
		tmp1 := v.Point().Mul(aNeg, gVec[n-1])
		tmp2 := v.Point().Mul(v.Scalar().Mul(aVecN, bNeg), h)

		P = v.Point().Add(v.Point().Add(tmp1, tmp2), P)
		n = n - 1
	}

	n1 := n / 2

	// step 3
	// Verifier receive L and R from Prover
	var L, R kyber.Point
	L, R = <-v.input2, <-v.input2

	// step 4
	// generate challenge value
	z := v.Scalar()
	zInv := v.Scalar().Inv(z)

	// step 5
	gVec1 := make([]kyber.Point, n1)
	for i := 0; i < n1; i++ {
		left1 := v.Point().Mul(zInv, gVec[:n1][i])
		right1 := v.Point().Mul(z, gVec[n1:][i])
		gVec1[i] = v.Point().Add(left1, right1)
	}
	z2 := v.Scalar().Mul(z, z)
	z2Inv := v.Scalar().Inv(z2)
	P1 := v.Point().Add(v.Point().Add(P, v.Point().Mul(z2, L)), v.Point().Mul(z2Inv, R))

	ret := v.RecursiveVerify(gVec1, h, P1, n1)
	return ret
}
