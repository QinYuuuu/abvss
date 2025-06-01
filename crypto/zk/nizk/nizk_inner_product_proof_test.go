package nizk

import (
	"testing"

	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/util/random"
)

func Test_NizkIPAProof_Marshal(t *testing.T) {
	n := int64(10)
	group := edwards25519.NewBlakeSHA256Ed25519()
	rand := random.New()
	params := SetupNizkIPA(group, n, rand)
	aVec := make([]kyber.Scalar, n)
	yVec := make([]kyber.Scalar, n)
	for i := int64(0); i < n; i++ {
		aVec[i] = group.Scalar().Pick(rand)
		yVec[i] = group.Scalar().Pick(rand)
	}
	proof, err := params.Prove(aVec, yVec)
	if err != nil {
		t.Error(err)
	}
	proofBytes, err := MarshalNizkIPAProofToBinary(proof)
	if err != nil {
		t.Error(err)
	}
	proof2, err := UnmarshalNizkIPAProofFromBinary(group, proofBytes)
	if err != nil {
		t.Error(err)
	}
	result, err := params.Verify(proof2, yVec)
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Error("verify failed")
	}
}
