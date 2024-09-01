package iipa

import "go.dedis.ch/kyber/v3"

type Prover struct {
	crs *CRS
	kyber.Group
}

func (p Prover) RecursiveProve(gVec []kyber.Point, h, P kyber.Point, aVec, bVec []kyber.Scalar, n int) {
	aLen := len(aVec)
	gLen := len(gVec)
	if aLen != len(bVec) {
		return
	}
	// step 1 Prover send aVec to Verifier
	if n == 1 {
		return
	}

	var proofStep []kyber.Scalar
	// step 2
	// Prover send aVec[n] to Verifier
	if n%2 == 1 {
		aNeg := p.Scalar().Neg(aVec[aLen-1])
		bNeg := p.Scalar().Neg(bVec[aLen-1])
		tmp1 := p.Point().Mul(aNeg, gVec[gLen-1])
		tmp2 := p.Point().Mul(p.Scalar().Mul(aVec[aLen-1], bNeg), h)

		P = p.Point().Add(p.Point().Add(tmp1, tmp2), P)
		proofStep = append(proofStep, aNeg, bNeg)
		n = n - 1
	}
	n1 := n / 2

	// step 3
	cl := p.Scalar().Zero()
	cr := p.Scalar().Zero()
	var L, R kyber.Point
	for i := 0; i < n1; i++ {
		tmp1 := p.Scalar().Mul(aVec[:n1][i], bVec[n1:][i])
		cl = p.Scalar().Add(cl, tmp1)
		tmp2 := p.Scalar().Mul(aVec[n1:][i], bVec[:n1][i])
		cr = p.Scalar().Add(cr, tmp2)
		L = p.Point().Mul(aVec[:n1][i], gVec[n1:][i])
		R = p.Point().Mul(aVec[n1:][i], gVec[:n1][i])
	}
	L = p.Point().Add(p.Point().Mul(cl, h), L)
	R = p.Point().Add(p.Point().Mul(cr, h), R)
	//p.RecursiveProve(n1)

	// z is the challenge value
	var z kyber.Scalar
	zInv := p.Scalar().Inv(z)
	// step 5, step 6
	gVec1 := make([]kyber.Point, n1)
	aVec1 := make([]kyber.Scalar, n1)
	bVec1 := make([]kyber.Scalar, n1)
	for i := 0; i < n1; i++ {
		left := p.Scalar().Mul(z, bVec[:n1][i])
		right := p.Scalar().Mul(zInv, bVec[n1:][i])
		bVec1[i] = p.Scalar().Add(left, right)

		left1 := p.Point().Mul(zInv, gVec[:n1][i])
		right1 := p.Point().Mul(z, gVec[n1:][i])
		gVec1[i] = p.Point().Add(left1, right1)

		left = p.Scalar().Mul(z, aVec[:n1][i])
		right = p.Scalar().Mul(zInv, aVec[n1:][i])
		aVec1[i] = p.Scalar().Add(left, right)
	}

	z2 := p.Scalar().Mul(z, z)
	z2Inv := p.Scalar().Inv(z2)

	P1 := p.Point().Add(p.Point().Add(P, p.Point().Mul(z2, L)), p.Point().Mul(z2Inv, R))
	p.RecursiveProve(gVec1, h, P1, aVec1, bVec1, n1)
	return
}
