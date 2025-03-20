package harts

import (
	"log/slog"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/commit/pedersen"
	"github.com/QinYuuuu/abvss/pkg"
	"go.dedis.ch/kyber/v3"
)

const (
	Commit string = "harts.Commit"
	Row    string = "harts.Row"
	Column string = "harts.Column"
	Vote   string = "harts.Vote"
	Done   string = "harts.Done"
)

type CommitPayload struct {
	CM      []kyber.Point
	row0_S  []kyber.Point
	row0_Pi []kyber.Point
}

type HAVSSNode struct {
	n, tc, tr int64
	p         *big.Int
	group     kyber.Group
	*dealer
	*party
}

type Share struct {
	XIndex      int
	PolyY       []*big.Int     // S(i,y)的单变量多项式
	Commitments *pedersen.Comm // 多项式承诺
	//Proofs      []NIZKProof    // NIZK证明
}

type dealer struct {
	biPoly *pkg.BivariatePoly
}

type party struct {
	voteSenders []bool
	doneSenders []bool
	voteCounter int64
	doneCounter int64
}

func (ss *HAVSSNode) DealerCommit() {
	if ss.dealer == nil {
		slog.Error("not dealer, cannot commit")
	}
	CPoly := make([]*pkg.Poly, ss.n)
	for i := range ss.n {
		CPoly[i] = ss.biPoly.EvalAtXMod(big.NewInt(int64(i)), ss.p)
	}
	var g kyber.Point
	S := make([]kyber.Point, ss.n)
	for i := range ss.n {
		Ci0 := CPoly[i].EvalMod(big.NewInt(0), ss.p).Int64()
		Ci0Scale := ss.group.Scalar().SetInt64(Ci0)
		S[i] = ss.group.Point().Mul(Ci0Scale, g)
	}
}

func (ss *HAVSSNode) DealerDistribute() map[int]*Share {
	if ss.dealer == nil {
		slog.Error("not dealer, cannot distribute")
	}
	shares := make(map[int]*Share)
	var i int64
	for i = 0; i <= ss.n; i++ {
	}
	return shares
}

func (ss *HAVSSNode) handleColumn(sender int64) {

}

func (ss *HAVSSNode) handleVote(sender int64) {
	if ss.voteSenders[sender] {
		slog.Error("have receive vote", slog.Any("fromID", sender))
	}
	ss.voteCounter++
	if ss.voteCounter >= ss.n-ss.tc {
		// send done
	}
}

func (ss *HAVSSNode) handleDone(sender int64) {
	if ss.doneSenders[sender] {
		slog.Error("have receive vote", slog.Any("fromID", sender))
	}
	ss.doneCounter++
	if ss.doneCounter >= ss.tc+1 {
		// send done
	}
	if ss.doneCounter >= ss.n-ss.tc {
		// terminate
	}
}
