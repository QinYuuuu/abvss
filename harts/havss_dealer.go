package harts

import (
	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v3"
	"log/slog"
	"math/big"
)

type dealer struct {
	biPoly *pkg.BivariatePoly
}

func (vss *HAVSSImpl) dealerCommit() {
	if vss.dealer == nil {
		slog.Error("not dealer, cannot commit")
	}
	CPoly := make([]*pkg.Poly, vss.n)
	for i := range vss.n {
		CPoly[i] = vss.biPoly.EvalAtXMod(big.NewInt(i), vss.p)
	}
	var g kyber.Point
	S := make([]kyber.Point, vss.n)
	for i := range vss.n {
		Ci0 := CPoly[i].EvalMod(big.NewInt(0), vss.p).Int64()
		Ci0Scale := vss.group.Scalar().SetInt64(Ci0)
		S[i] = vss.group.Point().Mul(Ci0Scale, g)
	}
}

func (vss *HAVSSImpl) dealerDistribute() map[int]*Share {
	if vss.dealer == nil {
		slog.Error("not dealer, cannot distribute")
	}
	shares := make(map[int]*Share)
	var i int64
	for i = 0; i <= vss.n; i++ {
	}
	return shares
}
