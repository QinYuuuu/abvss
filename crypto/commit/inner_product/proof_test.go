package inner_product

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/util/random"
)

func Test_Proof_Marshal(t *testing.T) {
	// Setup test parameters
	n := int64(13) // Use power of 2 for simplicity
	group := edwards25519.NewBlakeSHA256Ed25519()
	r := random.New()
	// Create a CRS
	crs := NewCRS(n, group, r)
	// Create test vectors of length n
	aVec := make([]kyber.Scalar, n)
	yVec := make([]kyber.Scalar, n)
	for i := int64(0); i < n; i++ {
		aVec[i] = group.Scalar().Pick(r)
		yVec[i] = group.Scalar().Pick(r)
	}
	// public: A, yVec, v
	// Prover private: aVec
	// Test the InnerProductProve function first
	v, A := crs.InnerProductProveInput(aVec, yVec)
	// Call the function under test
	proof, err := crs.NonInteractReduceProve(A, v, aVec, yVec)
	assert.NoError(t, err)
	// Verify that the results are not nil
	assert.NotNil(t, proof.aVecToSend)
	assert.NotNil(t, proof.lVec)
	assert.NotNil(t, proof.rVec)
	// Marshal the proof
	proofBytes, err := MarshalToBinary(proof)

	assert.NoError(t, err)
	assert.NotNil(t, proofBytes)
	// Unmarshal the proof
	unmarshaledProof, err := UnmarshalFromBinary(group, proofBytes)
	assert.NoError(t, err)
	assert.NotNil(t, unmarshaledProof)

	// Compare the unmarshaled proof with the original proof
	for i := 0; i < len(proof.lVec); i++ {
		assert.True(t, assert.Equal(t, proof.lVec[i].String(), unmarshaledProof.lVec[i].String()), "lVec[%d] should be equal", i)
	}
	for i := 0; i < len(proof.rVec); i++ {
		assert.True(t, assert.Equal(t, proof.rVec[i].String(), unmarshaledProof.rVec[i].String()), "rVec[%d] should be equal", i)
	}
	for i := 0; i < len(proof.aVecToSend); i++ {
		assert.True(t, assert.Equal(t, proof.aVecToSend[i].String(), unmarshaledProof.aVecToSend[i].String()), "aVecToSend[%d] should be equal", i)
	}
}
