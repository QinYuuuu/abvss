package nizk

import (
	"github.com/QinYuuuu/abvss/crypto/commit/inner_product"
	"github.com/QinYuuuu/abvss/pkg/protobuf"
	"go.dedis.ch/kyber/v4"
	"google.golang.org/protobuf/proto"
)

type NizkIPAProof struct {
	_A       kyber.Point    // _A = g[]^a[] * h^r
	_S       kyber.Point    // _S = g[]^sVec[] + h^v
	_T       kyber.Point    // _T = g^(sVec[] * y[])
	cVec     []kyber.Scalar // cVec = aVec[] + sVec[] * z
	v        kyber.Scalar   // v = r + z
	tHat     kyber.Scalar   // tHat = <cVec[], y[]>
	gExpC    kyber.Point    // g[]^cVec[]
	gExpU    kyber.Point
	ipaProof *inner_product.Proof
}

func MarshalNizkIPAProofToBinary(proof *NizkIPAProof) ([]byte, error) {
	proofMsg, err := MarshalNizkIPAProofToProto(proof)
	if err != nil {
		return nil, err
	}
	return proto.Marshal(proofMsg)
}

func UnmarshalNizkIPAProofFromBinary(group kyber.Group, data []byte) (*NizkIPAProof, error) {
	proofMsg := &protobuf.NizkIPAProof{}
	err := proto.Unmarshal(data, proofMsg)
	if err != nil {
		return nil, err
	}
	return UnmarshalNizkIPAProofFromProto(group, proofMsg)
}

func MarshalNizkIPAProofToProto(proof *NizkIPAProof) (*protobuf.NizkIPAProof, error) {
	ABytes, err := proof._A.MarshalBinary()
	if err != nil {
		return nil, err
	}
	SBytes, err := proof._S.MarshalBinary()
	if err != nil {
		return nil, err
	}
	TBytes, err := proof._T.MarshalBinary()
	if err != nil {
		return nil, err
	}
	cVecBytes := make([][]byte, len(proof.cVec))
	for i, c := range proof.cVec {
		cVecBytes[i], err = c.MarshalBinary()
		if err != nil {
			return nil, err
		}
	}
	vBytes, err := proof.v.MarshalBinary()
	if err != nil {
		return nil, err
	}
	tHatBytes, err := proof.tHat.MarshalBinary()
	if err != nil {
		return nil, err
	}
	gExpCBytes, err := proof.gExpC.MarshalBinary()
	if err != nil {
		return nil, err
	}
	gExpUBytes, err := proof.gExpU.MarshalBinary()
	if err != nil {
		return nil, err
	}
	ipaProof, err := inner_product.MarshalToProto(proof.ipaProof)
	if err != nil {
		return nil, err
	}
	return &protobuf.NizkIPAProof{
		A:        ABytes,
		S:        SBytes,
		T:        TBytes,
		CVec:     cVecBytes,
		V:        vBytes,
		THat:     tHatBytes,
		GExpC:    gExpCBytes,
		GExpU:    gExpUBytes,
		IpaProof: ipaProof,
	}, nil
}

func UnmarshalNizkIPAProofFromProto(group kyber.Group, proofMsg *protobuf.NizkIPAProof) (*NizkIPAProof, error) {
	A := group.Point()
	err := A.UnmarshalBinary(proofMsg.A)
	if err != nil {
		return nil, err
	}
	S := group.Point()
	err = S.UnmarshalBinary(proofMsg.S)
	if err != nil {
		return nil, err
	}
	T := group.Point()
	err = T.UnmarshalBinary(proofMsg.T)
	if err != nil {
		return nil, err
	}
	cVec := make([]kyber.Scalar, len(proofMsg.CVec))
	for i, c := range proofMsg.CVec {
		cVec[i] = group.Scalar()
		err = cVec[i].UnmarshalBinary(c)
		if err != nil {
			return nil, err
		}
	}
	v := group.Scalar()
	err = v.UnmarshalBinary(proofMsg.V)
	if err != nil {
		return nil, err
	}
	tHat := group.Scalar()
	err = tHat.UnmarshalBinary(proofMsg.THat)
	if err != nil {
		return nil, err
	}
	gExpC := group.Point()
	err = gExpC.UnmarshalBinary(proofMsg.GExpC)
	if err != nil {
		return nil, err
	}
	gExpU := group.Point()
	err = gExpU.UnmarshalBinary(proofMsg.GExpU)
	if err != nil {
		return nil, err
	}
	ipaProof, err := inner_product.UnmarshalFromProto(group, proofMsg.IpaProof)
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
		gExpC:    gExpC,
		gExpU:    gExpU,
		ipaProof: ipaProof,
	}, nil
}
