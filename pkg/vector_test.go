package pkg

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVecPow(t *testing.T) {
	v1 := []*big.Int{new(big.Int).SetInt64(4), new(big.Int).SetInt64(5)}
	v2 := []*big.Int{new(big.Int).SetInt64(3), new(big.Int).SetInt64(6)}
	p := new(big.Int).SetInt64(7)
	get, err := VecPow(v1, v2, p)
	assert.Nil(t, err, "err in vec pow")
	tmp1 := new(big.Int).Exp(v1[0], v2[0], p)
	tmp2 := new(big.Int).Exp(v1[1], v2[1], p)
	tmp3 := new(big.Int).Mul(tmp1, tmp2)
	want := new(big.Int).Mod(tmp3, p)
	assert.Equal(t, get, want, "vector power")
}

func TestMatrixMultiplyVector(t *testing.T) {
	tests := []struct {
		name    string
		matrix  [][]*big.Int
		vector  []*big.Int
		want    []*big.Int
		wantErr bool
	}{
		{
			name: "2x3 matrix with 3x1 vector",
			matrix: [][]*big.Int{
				{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
				{big.NewInt(4), big.NewInt(5), big.NewInt(6)},
			},
			vector:  []*big.Int{big.NewInt(7), big.NewInt(8), big.NewInt(9)},
			want:    []*big.Int{big.NewInt(50), big.NewInt(122)},
			wantErr: false,
		},
		{
			name: "1x1 matrix with 1x1 vector",
			matrix: [][]*big.Int{
				{big.NewInt(5)},
			},
			vector:  []*big.Int{big.NewInt(10)},
			want:    []*big.Int{big.NewInt(50)},
			wantErr: false,
		},
		{
			name: "3x2 matrix with 3x1 vector - dimension mismatch",
			matrix: [][]*big.Int{
				{big.NewInt(1), big.NewInt(2)},
				{big.NewInt(3), big.NewInt(4)},
				{big.NewInt(5), big.NewInt(6)},
			},
			vector:  []*big.Int{big.NewInt(7), big.NewInt(8), big.NewInt(9)},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty matrix",
			matrix:  [][]*big.Int{},
			vector:  []*big.Int{big.NewInt(1), big.NewInt(2)},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatrixMulVector(tt.matrix, tt.vector)

			// Check error status
			if (err != nil) != tt.wantErr {
				t.Errorf("MatrixMultiplyVector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Skip result check if we expected an error
			if tt.wantErr {
				return
			}

			// Check if result matches expected output
			if !CompareVectors(got, tt.want) {
				t.Errorf("MatrixMultiplyVector() = %v, want %v", got, tt.want)
			}
		})
	}
}
