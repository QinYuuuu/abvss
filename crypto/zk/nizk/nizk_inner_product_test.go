package nizk

import (
	"testing"

	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"gotest.tools/v3/assert"
)

func TestNIZKInnerProduct_Prove_and_Verify(t *testing.T) {
	n := int64(10)
	group := edwards25519.NewBlakeSHA256Ed25519()
	rand := group.RandomStream()
	param := SetupNizkIPA(group, n, rand)

	// generate test input
	aVec := make([]kyber.Scalar, n)
	yVec := make([]kyber.Scalar, n)
	for i := int64(0); i < n; i++ {
		aVec[i] = group.Scalar().Pick(rand)
		yVec[i] = group.Scalar().Pick(rand)
	}
	proof, comm, err := param.Prove(aVec, yVec)
	assert.NilError(t, err)

	result, err := param.Verify(proof, comm, yVec)
	assert.NilError(t, err)
	assert.Equal(t, result, true)
}
