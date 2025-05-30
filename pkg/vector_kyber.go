package pkg

import (
	"errors"
	"fmt"

	"go.dedis.ch/kyber/v4"
)

func DotProductExpKyber(g []kyber.Point, a []kyber.Scalar) (kyber.Point, error) {
	if len(g) != len(a) {
		return nil, errors.New("the input length is different")
	}
	if len(g) == 0 {
		return nil, errors.New("vectors cannot be empty")
	}
	// Initialize result with zero value
	result := g[0].Clone().Mul(a[0], g[0])
	for i := 1; i < len(g); i++ {
		tmp := g[i].Clone().Mul(a[i], g[i])
		result = result.Add(result, tmp)
	}
	return result, nil
}

// DotProductKyber calculates the dot product of two vectors of Scalars
func DotProductKyber(v1, v2 []kyber.Scalar) (kyber.Scalar, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}

	// Ensure we have a valid scalar implementation to create our dot product result
	if len(v1) == 0 {
		return nil, errors.New("vectors cannot be empty")
	}

	// Initialize result with zero value
	dot := v1[0].Clone().Zero()

	for i := 0; i < len(v1); i++ {
		// temp = v1[i] * v2[i]
		temp := v1[i].Clone().Mul(v1[i], v2[i])
		// dot += temp
		dot.Add(dot, temp)
	}

	return dot, nil
}

// VecPowKyber computes the product of v1[i]^v2[i] mod m for all i
func VecPowKyber(v1, v2 []kyber.Scalar, m kyber.Scalar) (kyber.Scalar, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}

	if len(v1) == 0 {
		return nil, errors.New("vectors cannot be empty")
	}

	// Initialize result with one value (multiplicative identity)
	dot := v1[0].Clone().One()

	for i := 0; i < len(v1); i++ {
		// Kyber doesn't have a direct Exp method like big.Int
		// We need to use specialized implementation based on your needs
		// This is a simplified version assuming we have a function ExpMod:
		// temp = v1[i]^v2[i] mod m
		temp := ExpMod(v1[i], v2[i], m)
		// dot *= temp
		dot.Mul(dot, temp)
	}

	return dot, nil
}

// ExpMod calculates base^exp mod m for Scalar values
// Note: Kyber doesn't provide a direct Exp function - this is a placeholder
// You'd need to implement this based on the specific Scalar implementation
func ExpMod(base, exp, mod kyber.Scalar) kyber.Scalar {
	// Placeholder implementation - you'll need to replace this with actual modular exponentiation
	// This could involve converting to big.Int, performing operation, and converting back
	// Or using a specialized implementation for the specific curve being used
	result := base.Clone().One()

	// This is just a sketch - actual implementation depends on your Scalar implementation
	// For certain groups, you might need to implement your own exponentiation algorithm

	return result
}

// VecAddKyber returns v1 + v2
func VecAddKyber(v1, v2 []kyber.Scalar) ([]kyber.Scalar, error) {
	if len(v1) != len(v2) {
		return nil, errors.New("the input length is different")
	}

	v := make([]kyber.Scalar, len(v1))
	for i := 0; i < len(v); i++ {
		v[i] = v1[i].Clone()
		v[i].Add(v[i], v2[i])
	}

	return v, nil
}

// MatrixMulVectorKyber multiplies a matrix by a vector
func MatrixMulVectorKyber(matrix [][]kyber.Scalar, vector []kyber.Scalar) ([]kyber.Scalar, error) {
	// Check if matrix dimensions match the vector length
	if len(matrix) == 0 || len(matrix[0]) != len(vector) {
		return nil, fmt.Errorf("matrix columns must equal vector length")
	}

	// Initialize result vector
	result := make([]kyber.Scalar, len(matrix))
	for i := range result {
		result[i] = vector[0].Clone().Zero() // Clone first element to get correct implementation
	}

	// Perform matrix-vector multiplication
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(vector); j++ {
			// Calculate matrix element * vector element
			temp := matrix[i][j].Clone().Mul(matrix[i][j], vector[j])
			// Add product to corresponding position in result vector
			result[i].Add(result[i], temp)
		}
	}

	return result, nil
}

// CompareVectorsKyber checks if two Scalar vectors are equal
func CompareVectorsKyber(a, b []kyber.Scalar) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		// Create a scalar containing the difference
		diff := a[i].Clone().Sub(a[i], b[i])
		// If the difference is not zero, vectors are not equal
		if !diff.Equal(diff.Clone().Zero()) {
			return false
		}
	}

	return true
}

func VecScalarMulKyber(vec []kyber.Scalar, s kyber.Scalar) []kyber.Scalar {
	result := make([]kyber.Scalar, len(vec))
	for i := range vec {
		result[i] = vec[i].Clone().Mul(vec[i], s)
	}
	return result
}
