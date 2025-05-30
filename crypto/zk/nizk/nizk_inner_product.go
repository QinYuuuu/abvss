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

type NizkIPAProof struct {
	_S       kyber.Point    // _S = g[]^sVec[] + h^v
	_T       kyber.Point    // _T = g^(sVec[] * y[])
	cVec     []kyber.Scalar // cVec = aVec[] + sVec[] * z
	v        kyber.Scalar   // v = r + z
	tHat     kyber.Scalar   // tHat = <cVec[], y[]>
	gExpC    kyber.Point    // g[]^cVec[]
	gExpU    kyber.Point
	ipaProof *inner_product.Proof
}

func SetupNizkIPA(group kyber.Group, n int64, rand cipher.Stream) *NizkIPAParam {
	params := &NizkIPAParam{
		g:   group.Point().Base(),
		crs: inner_product.NewCRS(n, group, rand),
	}
	return params
}

// g[]^a[] * h^r
func (param *NizkIPAParam) Prove(aVec []kyber.Scalar) (*NizkIPAProof, kyber.Point, error) {
	group := param.crs.GetGroup()
	yVec := param.crs.GetY()
	rand := param.crs.GetRand()
	r, A := param.crs.InnerProductProveInput(aVec)
	gExpU := group.Point().Mul(r, param.g)

	sVec := make([]kyber.Scalar, param.crs.GetN())
	for i := int64(0); i < param.crs.GetN(); i++ {
		sVec[i] = group.Scalar().Pick(rand)
	}
	rho := group.Scalar().Pick(rand)

	power, err := pkg.DotProductKyber(sVec, param.crs.GetY())
	if err != nil {
		return nil, nil, err
	}
	T := group.Point().Mul(power, param.g)
	S, err := pkg.DotProductExpKyber(param.crs.GetG(), sVec)
	if err != nil {
		return nil, nil, err
	}
	S = group.Point().Add(S, group.Point().Mul(rho, param.crs.GetH()))

	z := param.generateRandom(S, T)
	cVec, err := pkg.VecAddKyber(aVec, pkg.VecScalarMulKyber(sVec, z))
	v := group.Scalar().Add(r, group.Scalar().Mul(rho, z))
	tHat, err := pkg.DotProductKyber(cVec, yVec)
	if err != nil {
		return nil, nil, err
	}
	C, err := pkg.DotProductExpKyber(param.crs.GetG(), cVec)
	if err != nil {
		return nil, nil, err
	}

	proof, err := param.crs.NonInteractReduceProve(A, r, aVec)
	if err != nil {
		return nil, nil, err
	}

	return &NizkIPAProof{
		_S:       S,
		_T:       T,
		cVec:     cVec,
		v:        v,
		tHat:     tHat,
		gExpC:    C,
		gExpU:    gExpU,
		ipaProof: proof,
	}, A, nil
}

func (param *NizkIPAParam) Verify(proof *NizkIPAProof, A kyber.Point) (bool, error) {
	S := proof._S
	T := proof._T

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
		tHatWant, err := pkg.DotProductKyber(proof.cVec, param.crs.GetY())
		if err != nil {
			return false, err
		}
		if !(tHatWant.String() == proof.tHat.String()) {
			return false, nil
		}
	}
	{
		if !param.crs.NonInteractVerify(proof.ipaProof, A) {
			return false, nil
		}
	}
	return true, nil
}

func (param *NizkIPAParam) generateRandom(S, T kyber.Point) kyber.Scalar {
	LBytes := []byte(S.String())
	RBytes := []byte(T.String())
	zByte := append(LBytes, RBytes...)
	group := param.crs.GetGroup()
	z := group.Scalar().SetBytes(zByte)
	return z
}
