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

	}
	return nil
}
