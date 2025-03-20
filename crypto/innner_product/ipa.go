package innnerproduct

import "math/big"

func innerProduct(a, b []*big.Int, p *big.Int) *big.Int {
	result := big.NewInt(0)
	for i := range a {
		tmp := new(big.Int).Mul(a[i], b[i])
		result.Add(result, tmp)
	}

	return result.Mod(result, p)
}
