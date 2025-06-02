package pkg

import (
	"testing"

	"go.dedis.ch/kyber/v4/group/edwards25519"
)

func Test_Interpolation_Exp_At_Index_Kyber(t *testing.T) {
	// generate random data
	group := edwards25519.NewBlakeSHA256Ed25519()

	scalar := group.Scalar().Pick(group.RandomStream())
	point := group.Point().Mul(scalar, nil)

	scalar2 := group.Scalar().Pick(group.RandomStream())
}
