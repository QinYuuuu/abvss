package iipa

import (
	"github.com/stretchr/testify/assert"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/util/random"
	"testing"
)

func TestNewCRS(t *testing.T) {
	// Setup test parameters
	n := int64(5)
	group := edwards25519.NewBlakeSHA256Ed25519()
	r := random.New()
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
	assert.Equal(t, int(n), len(crs.g))

	// Verify that y array has correct length
	assert.Equal(t, int(n), len(crs.y))

	// Verify that g elements are not nil and different
	for i := int64(0); i < n; i++ {
		assert.NotNil(t, crs.g[i])

		// Each g[i] should be a valid point in the group
		assert.True(t, crs.g[i].Equal(crs.g[i]))

		// Check that elements are likely different
		if i > 0 {
			assert.False(t, crs.g[i].Equal(crs.g[i-1]), "Generated points in g should be different")
		}
	}

	// Verify that h is not nil
	assert.NotNil(t, crs.h)
	assert.True(t, crs.h.Equal(group.Point().Set(crs.h)))

	// Verify that y elements are not nil and within range [0, p)
	for i := int64(0); i < n; i++ {
		assert.NotNil(t, crs.y[i])
	}

	// Create another CRS with the same parameters
	// They should be different due to randomness
	crs2 := NewCRS(n, group, r)

	differentPoints := false
	for i := int64(0); i < n; i++ {
		if !crs.g[i].Equal(crs2.g[i]) {
			differentPoints = true
			break
		}
	}

	assert.True(t, differentPoints, "CRS generation should produce different points with the same parameters")
}

func TestNonInteractProve(t *testing.T) {
	// Setup test parameters
	n := int64(2) // Use power of 2 for simplicity
	group := edwards25519.NewBlakeSHA256Ed25519()
	r := random.New()

	// Create a CRS
	crs := NewCRS(n, group, r)

	// Create test vectors of length n
	aVec := make([]kyber.Scalar, n)
	for i := int64(0); i < n; i++ {
		aVec[i] = group.Scalar().Pick(r)
	}
	//z := utils.RandomNum(p)
	// Test the InnerProductProve function first
	v, A := crs.InnerProductProveInput(aVec)
	//A1 := crs.InnerProductProve(A, v, z)
	// Now test NonInteractProve
	// We use the same parameters as the CRS
	gVec := crs.g
	h := crs.h
	yVec := crs.y

	// Call the function under test
	resultAVec, LVec, RVec := crs.NonInteractProve(gVec, h, A, v, aVec, yVec, int(n))

	// Verify that the results are not nil
	assert.NotNil(t, resultAVec)
	assert.NotNil(t, LVec)
	assert.NotNil(t, RVec)
	assert.Equal(t, 1, len(resultAVec))

	// The number of L and R values should be log2(n) if n is a power of 2
	expectedProofSize := 0
	if isPowerOfTwo(int(n)) {
		tmp := int(n)
		for tmp > 1 {
			tmp /= 2
			expectedProofSize++
		}
	}
	result := crs.NonInteractVerify(gVec, LVec, RVec, h, A, resultAVec, yVec, int(n))
	assert.True(t, result)
}

// Helper function to check if a number is a power of 2
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}
