package nizk

import (
	"crypto/cipher"

	"github.com/QinYuuuu/abvss/crypto/commit/inner_product"
	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v4"
)

type NizkIPAParam struct {
	g   kyber.Point
	crs *inner_product.CRS
}

func SetupNizkIPA(group kyber.Group, n int64, rand cipher.Stream) *NizkIPAParam {
	params := &NizkIPAParam{
		g:   group.Point().Base(),
		crs: inner_product.NewCRS(n, group, rand),
	}
	return params
}

// g[]^a[] * h^r
func (param *NizkIPAParam) Prove(aVec, yVec []kyber.Scalar) (*NizkIPAProof, error) {
	group := param.crs.GetGroup()
	rand := param.crs.GetRand()
	r, A := param.crs.InnerProductProveInput(aVec, yVec)
	gExpU := group.Point().Mul(r, param.g)

	sVec := make([]kyber.Scalar, param.crs.GetN())
	for i := int64(0); i < param.crs.GetN(); i++ {
		sVec[i] = group.Scalar().Pick(rand)
	}
	rho := group.Scalar().Pick(rand)

	power, err := pkg.DotProductKyber(sVec, yVec)
	if err != nil {
		return nil, err
	}
	T := group.Point().Mul(power, param.g)
	S, err := pkg.DotProductExpKyber(param.crs.GetG(), sVec)
	if err != nil {
		return nil, err
	}
	S = group.Point().Add(S, group.Point().Mul(rho, param.crs.GetH()))

	z := param.generateRandom(S, T)
	cVec, err := pkg.VecAddKyber(aVec, pkg.VecScalarMulKyber(sVec, z))
	v := group.Scalar().Add(r, group.Scalar().Mul(rho, z))
	tHat, err := pkg.DotProductKyber(cVec, yVec)
	if err != nil {
		return nil, err
	}
	C, err := pkg.DotProductExpKyber(param.crs.GetG(), cVec)
	if err != nil {
		return nil, err
	}

	proof, err := param.crs.NonInteractReduceProve(A, r, aVec, yVec)
	if err != nil {
		return nil, err
	}

	return &NizkIPAProof{
		_A:       A,
		_S:       S,
		_T:       T,
		cVec:     cVec,
		v:        v,
		tHat:     tHat,
		gExpC:    C,
		gExpU:    gExpU,
		ipaProof: proof,
	}, nil
}

func (param *NizkIPAParam) ProveForPoly(aVec []kyber.Scalar, xIndex kyber.Scalar) (*NizkIPAProof, error) {
	group := param.crs.GetGroup()
	yVec := make([]kyber.Scalar, len(aVec))
	for j := 0; j < len(aVec); j++ {
		if j == 0 {
			yVec[j] = group.Scalar().One()
			continue
		}
		yVec[j] = group.Scalar().Mul(xIndex, yVec[j-1])
	}
	return param.Prove(aVec, yVec)
}

func (param *NizkIPAParam) Verify(proof *NizkIPAProof, yVec []kyber.Scalar) (bool, error) {
	S := proof._S
	T := proof._T
	A := proof._A
	z := param.generateRandom(S, T)
	group := param.crs.GetGroup()
	{
		// verify tHat = g^u * T^z
		left := group.Point().Mul(proof.tHat, param.g)
		right := group.Point().Mul(z, T)
		right = group.Point().Add(proof.gExpU, right)
		if !(left.String() == right.String()) {
			return false, nil
		}
	}
	{
		// verify g[]^cVec[] * h^v = S^z * A
		left, err := pkg.DotProductExpKyber(param.crs.GetG(), proof.cVec)
		if err != nil {
			return false, err
		}
		left = group.Point().Add(left, group.Point().Mul(proof.v, param.crs.GetH()))
		right := group.Point().Mul(z, S)
		right = group.Point().Add(A, right)
		if !(left.String() == right.String()) {
			return false, nil
		}
	}
	{
		// verify tHat = <cVec[], y[]>
		tHatWant, err := pkg.DotProductKyber(proof.cVec, yVec)
		if err != nil {
			return false, err
		}
		if !(tHatWant.String() == proof.tHat.String()) {
			return false, nil
		}
	}
	{
		if !param.crs.NonInteractVerify(proof.ipaProof, A, yVec) {
			return false, nil
		}
	}
	return true, nil
}

func (param *NizkIPAParam) VerifyForPoly(proof *NizkIPAProof, xIndex kyber.Scalar) (bool, error) {
	group := param.crs.GetGroup()
	yVec := make([]kyber.Scalar, len(proof.cVec))
	for j := 0; j < len(proof.cVec); j++ {
		if j == 0 {
			yVec[j] = group.Scalar().One()
			continue
		}
		yVec[j] = group.Scalar().Mul(xIndex, yVec[j-1])
	}
	return param.Verify(proof, yVec)
}

func (param *NizkIPAParam) generateRandom(S, T kyber.Point) kyber.Scalar {
	LBytes := []byte(S.String())
	RBytes := []byte(T.String())
	zByte := append(LBytes, RBytes...)
	group := param.crs.GetGroup()
	z := group.Scalar().SetBytes(zByte)
	return z
}
