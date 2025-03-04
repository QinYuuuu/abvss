package abdkg

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
)

func IIPA_Prover1(b int) (kyber.Point, kyber.Point) {
	curve := curve25519.NewBlakeSHA256Curve25519(true)
	product := curve.Scalar().SetInt64(0)
	s := make([]kyber.Scalar, b)
	for i := 0; i < b; i++ {
		s[i] = curve.Scalar()
		s[i] = curve.Scalar().Mul(s[i], s[i])
		product = curve.Scalar().Add(product, s[i])
	}
	g := curve.Point()
	T := curve.Point().Mul(product, g)
	S := curve.Point().Mul(curve.Scalar().SetInt64(0), g)
	for i := 0; i < b; i++ {
		tmp := curve.Point().Mul(s[i], g)
		S = curve.Point().Add(S, tmp)
	}
	return S, T
}

/*
func IIPA_Prover2(b int) (kyber.Point, kyber.Point) {
	curve := curve25519.NewBlakeSHA256Curve25519(true)
	product := curve.Scalar().SetInt64(0)
	c := make([]kyber.Scalar, b)
	z := curve.Scalar().Pick(curve.RandomStream())
	for i := 0; i < b; i++ {
		tmp := curve.Scalar().Mul(z, curve.Scalar().Pick(curve.RandomStream()))
		c[i] = curve.Scalar().Add(curve.Scalar().Pick(curve.RandomStream()), tmp)
	}
	g := curve.Point()
	T := curve.Point().Mul(product, g)
	S := curve.Point().Mul(curve.Scalar().SetInt64(0), g)
	for i := 0; i < b; i++ {
		tmp := curve.Point().Mul(s[i], g)
		S = curve.Point().Add(S, tmp)
	}
	return S, T
}*/
