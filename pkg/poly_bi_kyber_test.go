package pkg

import (
	"fmt"
	"testing"

	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func Test_BivariatePolyKyberImpl_EvalAtXMod(t *testing.T) {
	group := edwards25519.NewBlakeSHA256Ed25519()
	poly, err := NewBiPolyKyber(1, 1, group)
	if err != nil {
		t.Fatalf("NewBiPolyKyber failed: %v", err)
	}
	poly.SetCoefficient(0, 0, group, 1)
	poly.SetCoefficient(1, 0, group, 2)
	poly.SetCoefficient(0, 1, group, 3)
	poly.SetCoefficient(1, 1, group, 4)
	fmt.Println(poly.ToString())

	poly0 := poly.EvalAtXMod(group.Scalar().Zero())
	fmt.Println(poly0.ToString())
}
