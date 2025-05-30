package inner_product

import (
	"math/big"

	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v4"
)

func (crs *CRS) RecursiveVerify(gVec, LVec, RVec []kyber.Point, h, P kyber.Point, aVec, yVec, zVec []*big.Int, n int) bool {
	// step 1
	// 1.1 Verifier receive aVec from Prover (length of aVec == 1)
	// 1.2 verify P
	if n == 1 {
		// get from input channel
		aScalar := crs.group.Scalar().SetInt64(aVec[0].Int64())
		left := crs.group.Point().Mul(aScalar, gVec[0])
		exp := new(big.Int).Mul(aVec[0], yVec[0])
		right := crs.group.Point().Mul(crs.group.Scalar().SetInt64(exp.Int64()), h)
		PWant := crs.group.Point().Add(left, right)
		return P.Equal(PWant)
	}

	// step 2
	// 2.1 Verifier receive aVec[n], bVec[n] from Prover
	// 2.3 update P, aVec, gVec, n
	if n%2 == 1 {
		// get from input channel
		aVecN := crs.group.Scalar().SetInt64(aVec[0].Int64())
		yVecN := crs.group.Scalar().SetInt64(yVec[0].Int64())
		aNeg := crs.group.Scalar().Neg(aVecN)
		yNeg := crs.group.Scalar().Neg(yVecN)
		tmp1 := crs.group.Point().Mul(aNeg, gVec[n-1])
		tmp2 := crs.group.Point().Mul(crs.group.Scalar().Mul(aVecN, yNeg), h)
		P = crs.group.Point().Add(crs.group.Point().Add(tmp1, tmp2), P)
		n = n - 1
	}
	n1 := n / 2

	// step 3
	// Verifier receive L and R from Prover
	L, R := LVec[0], RVec[0]

	// step 4
	// generate challenge value
	zScalar := crs.group.Scalar().SetInt64(zVec[0].Int64())
	zInv := new(big.Int).Neg(zVec[0])
	zScalarInv := crs.group.Scalar().SetInt64(zInv.Int64())

	// step 5
	gVec1 := make([]kyber.Point, n1)
	for i := 0; i < n1; i++ {
		left1 := crs.group.Point().Mul(zScalarInv, gVec[:n1][i])
		right1 := crs.group.Point().Mul(zScalar, gVec[n1:][i])
		gVec1[i] = crs.group.Point().Add(left1, right1)
	}
	z2 := crs.group.Scalar().Mul(zScalar, zScalar)
	z2Inv := crs.group.Scalar().Inv(z2)
	Lz2 := crs.group.Point().Mul(z2, L)
	Rz2Inv := crs.group.Point().Mul(z2Inv, R)
	P1 := crs.group.Point().Add(crs.group.Point().Add(P, Lz2), Rz2Inv)

	ret := crs.RecursiveVerify(gVec1, LVec[1:], RVec[1:], h, P1, aVec[1:], yVec[1:], zVec[1:], n1)
	return ret
}

func (crs *CRS) NonInteractVerify(proof *Proof, P kyber.Point) bool {
	// copy from common reference string
	gVec := make([]kyber.Point, len(crs.gVec))
	copy(gVec, crs.gVec)
	h := crs.h
	yVec := make([]kyber.Scalar, len(crs.yVec))
	copy(yVec, crs.yVec)
	n := crs.n

	// copy from proof
	aVec := proof.aVecToSend
	LVec := proof.lVec
	RVec := proof.rVec

	for n > 1 {
		// step 2
		// 2.1 Verifier receive aVec[n], bVec[n] from Prover
		// 2.3 update P, aVec, gVec, n
		if n%2 == 1 {
			a := aVec[len(aVec)-1]
			// slog.Info("Verifier	", slog.Any("n", n), slog.Any("a from prover", a.String()))
			aNeg := crs.group.Scalar().Neg(a)
			y := yVec[len(yVec)-1]
			// slog.Info("Verifier	", slog.Any("n", n), slog.String("y[-1]", y.String()))
			tmp1 := crs.group.Point().Mul(aNeg, gVec[len(gVec)-1])
			tmp2 := crs.group.Point().Mul(crs.group.Scalar().Mul(aNeg, y), h)
			P = crs.group.Point().Add(crs.group.Point().Add(tmp1, tmp2), P)
			// slog.Info("Verifier	", slog.Any("n", n), slog.String("p in n is odd", P.String()))
			n = n - 1
			aVec = aVec[:len(aVec)-1]
		}
		n1 := n / 2
		// step 3
		// Verifier receive L and R from Prover
		L, R := LVec[0], RVec[0]
		LBytes := []byte(L.String())
		RBytes := []byte(R.String())
		zByte := append(LBytes, RBytes...)

		// step 4
		// generate challenge value
		z := crs.group.Scalar().SetBytes(zByte)
		// slog.Info("Verifier nonInteractVerify z marshal", slog.Any("n", n), slog.Any("z", z.String()))
		zInv := crs.group.Scalar().Inv(z)
		// step 5
		gVec1 := make([]kyber.Point, n1)
		yVec1 := make([]kyber.Scalar, n1)
		for i := int64(0); i < n1; i++ {
			left := crs.group.Scalar().Mul(zInv, yVec[:n1][i])
			right := crs.group.Scalar().Mul(z, yVec[n1:][i])
			yVec1[i] = crs.group.Scalar().Add(left, right)

			left1 := crs.group.Point().Mul(zInv, gVec[:n1][i])
			right1 := crs.group.Point().Mul(z, gVec[n1:][i])
			gVec1[i] = crs.group.Point().Add(left1, right1)
		}
		z2 := crs.group.Scalar().Mul(z, z)
		z2Inv := crs.group.Scalar().Inv(z2)
		Lz2 := crs.group.Point().Mul(z2, L)
		Rz2Inv := crs.group.Point().Mul(z2Inv, R)
		P1 := crs.group.Point().Add(crs.group.Point().Add(P, Lz2), Rz2Inv)

		gVec = gVec1
		yVec = yVec1
		LVec = LVec[1:]
		RVec = RVec[1:]
		P = P1
		n = n1
	}
	// step 1
	// 1.1 Verifier receive aVec from Prover (length of aVec == 1)
	// 1.2 verify P
	left, _ := pkg.DotProductExpKyber(gVec, aVec)
	exp, _ := pkg.DotProductKyber(aVec, yVec)
	right := crs.group.Point().Mul(exp, h)
	PWant := crs.group.Point().Add(left, right)
	/*slog.Info("Verifier nonInteractVerify", slog.Any("verify final p", P.String() == PWant.String()))
	if P.String() != PWant.String() {
		slog.Info("Verifier nonInteractVerify", slog.Any("verify final p", P.String()), slog.Any("verify final p want", PWant.String()))
	}*/
	return P.Equal(PWant)
}
