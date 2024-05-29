package protocol

import (
	"errors"
	"math/big"
)

func (vss *ABVSS) VerifyLCM(lcm []*big.Int, index int) error {
	if vss.ABVSSV == nil {
		return errors.New("not a verifier")
	}
	if len(vss.jlist) == 0 {
		tuple := struct {
			index int
			lcm   []*big.Int
		}{index, lcm}
		vss.ilist = append(vss.ilist, tuple)
	} else {
		vss.jlist = append(vss.jlist, index)
	}
	return nil
}
