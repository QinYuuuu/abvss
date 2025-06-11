package pkg

import (
	"fmt"
	"testing"

	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func Test_Generate_Vandermonde(t *testing.T) {
	suite := edwards25519.NewBlakeSHA256Ed25519()

	// Define matrix dimensions
	n := 3 // Number of rows
	l := 4 // Number of columns

	fmt.Printf("Generating a %d x %d Vandermonde matrix...\n", n, l)
	vandermondeMatrix, err := GenerateVandermondeKyber(n, l, suite)
	if err != nil {
		fmt.Printf("Error generating matrix: %v\n", err)
		return
	}

	fmt.Println("Generated Vandermonde Matrix:")
	PrintKyberMatrix(vandermondeMatrix)
}
