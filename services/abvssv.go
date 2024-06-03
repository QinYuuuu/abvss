package services

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
	"math/big"
)

func (vss *ABVSS) VerifyInit() {
	vss.ABVSSV = &ABVSSV{
		count: 0,
		ilist: make([]struct {
			index int
			lcm   []*big.Int
		}, 0),
		jlist: make([]int, 0),
	}
}

func (vss *ABVSS) VerifyLCM(lcm []*big.Int, index int) error {
	if vss.ABVSSV == nil {
		return errors.New("not a verifier")
	}

	tuple := struct {
		index int
		lcm   []*big.Int
	}{index, lcm}
	vss.ilist = append(vss.ilist, tuple)
	vss.count++

	if vss.count == vss.nodenum-vss.degree && !vss.done {
		//log.Printf("node %v verify", vss.nodeid)
		xlist := make([]*big.Int, vss.nodenum-vss.degree)
		for i := 0; i < vss.degree+1; i++ {
			xlist[i] = new(big.Int).SetInt64(int64(vss.ilist[i].index + 1))
		}
		for i := 0; i < vss.vnum; i++ {
			ylist := make([]*big.Int, vss.nodenum-vss.degree)
			for j := 0; j < vss.nodenum-vss.degree; j++ {
				ylist[j] = vss.ilist[j].lcm[i]
			}
			poly, _ := polynomial.LagrangeInterpolation(xlist[:vss.degree+1], ylist[:vss.degree+1], vss.p)
			for i := vss.degree + 1 + 1; i < vss.nodenum-vss.degree; i++ {
				poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), vss.p)
				//log.Printf("verify %v \n", tmp.Cmp(ylist[i]) == 0)
			}
		}
		vss.done = true
	}
	return nil
}
