package inner_product

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v4"
	"go.dedis.ch/kyber/v4/group/edwards25519"
	"go.dedis.ch/kyber/v4/util/random"
)

func TestNewCRS(t *testing.T) {
	// Setup test parameters
	n := int64(5)
	group := edwards25519.NewBlakeSHA256Ed25519()
	r := group.RandomStream()
	// Assuming GetPrime returns a big.Int of given bit length

	// Create a CRS
	crs := NewCRS(n, group, r)

	// Verify that the CRS is not nil
	assert.NotNil(t, crs)

	// Check that the parameters are set correctly
	assert.Equal(t, n, crs.n)
	assert.Equal(t, group, crs.group)
	assert.Equal(t, r, crs.r)

	// Verify that g array has correct length
	assert.Equal(t, int(n), len(crs.gVec))

	// Verify that g elements are not nil and different
	for i := int64(0); i < n; i++ {
		assert.NotNil(t, crs.gVec[i])

		// Each g[i] should be a valid point in the group
		assert.True(t, crs.gVec[i].Equal(crs.gVec[i]))

		// Check that elements are likely different
		if i > 0 {
			assert.False(t, crs.gVec[i].Equal(crs.gVec[i-1]), "Generated points in g should be different")
		}
	}

	// Verify that h is not nil
	assert.NotNil(t, crs.h)
	assert.True(t, crs.h.Equal(group.Point().Set(crs.h)))

	// Create another CRS with the same parameters
	// They should be different due to randomness
	crs2 := NewCRS(n, group, r)

	differentPoints := false
	for i := int64(0); i < n; i++ {
		if !crs.gVec[i].Equal(crs2.gVec[i]) {
			differentPoints = true
			break
		}
	}
	assert.True(t, differentPoints, "CRS generation should produce different points with the same parameters")
}

func Test_NonInteract_Verify(t *testing.T) {
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

	// Verify that the results are not nil
	assert.NotNil(t, proof.aVecToSend)
	assert.NotNil(t, proof.lVec)
	assert.NotNil(t, proof.rVec)
	assert.Nil(t, err)

	result := crs.NonInteractVerify(proof, A, yVec)
	assert.True(t, result)
}
