package utils

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
