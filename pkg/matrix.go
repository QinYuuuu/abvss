package pkg

import (
	"fmt"

	"go.dedis.ch/kyber/v4"
)

// GenerateVandermondeKyber generates an n x l Vandermonde matrix with kyber.Scalar elements.
// The base scalar for row i (0-indexed) will be the kyber.Scalar representation of (i+1).
// You could also choose to generate random base scalars if needed.
func GenerateVandermondeKyber(row, column int, suite kyber.Group) ([][]kyber.Scalar, error) {
	if row <= 0 {
		return nil, fmt.Errorf("number of rows n must be positive, got %d", row)
	}
	if column <= 0 {
		return nil, fmt.Errorf("number of columns l must be positive, got %d", column)
	}

	matrix := make([][]kyber.Scalar, row)

	for i := 0; i < row; i++ {
		matrix[i] = make([]kyber.Scalar, column)

		// Determine the base scalar for this row.
		// For simplicity, we use the scalar representation of (i+1).
		// You could also use suite.Scalar().Pick(random.New()) for random base scalars.
		baseScalarForRow := suite.Scalar().SetInt64(int64(i + 1))

		// currentPower will hold baseScalarForRow^j
		currentPower := suite.Scalar().One() // Starts at baseScalarForRow^0

		for j := 0; j < column; j++ {
			// matrix[i][j] = baseScalarForRow^j
			matrix[i][j] = currentPower.Clone() // Store a copy

			// Prepare for the next iteration: currentPower * baseScalarForRow
			if j < column-1 { // No need to multiply after the last element
				currentPower.Mul(currentPower, baseScalarForRow)
			}
		}
	}

	return matrix, nil
}

// Helper function to print the matrix (scalars are printed in their string representation)
func PrintKyberMatrix(matrix [][]kyber.Scalar) {
	if matrix == nil || len(matrix) == 0 {
		fmt.Println("Empty matrix")
		return
	}
	rows := len(matrix)
	cols := len(matrix[0])

	for i := 0; i < rows; i++ {
		if len(matrix[i]) != cols {
			fmt.Printf("Row %d has inconsistent column count\n", i)
			continue
		}
		fmt.Printf("[")
		for j := 0; j < cols; j++ {
			// Using String() for readability. For storage, MarshalBinary might be preferred.
			fmt.Printf("%s", matrix[i][j].String())
			if j < cols-1 {
				fmt.Printf(", ")
			}
		}
		fmt.Println("]")
	}
}
