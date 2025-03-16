package harts

import "math/big"

type HAVSS struct {
}

type Share struct {
	XIndex      int
	PolyY       []*big.Int    // S(i,y)的单变量多项式
	Commitments *PedersenComm // 多项式承诺
	Proofs      []NIZKProof   // NIZK证明
}

type dealer struct {
}

func (ss *HAVSS) DealerCommit() {

}

func (ss *HAVSS) DealerDistribute() map[int]*Share {
	shares := make(map[int]*Share)
	for i := 1; i <= n; i++ {
		share := &Share{
			XIndex: i,
			PolyY:  evaluatePolynomialAtX(i),
			Proofs: generateNIZKProofs(i),
		}
		shares[i] = share
	}
	return shares
}
