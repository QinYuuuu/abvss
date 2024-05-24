package abvss

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
	"math/big"
)

func (vss *ABVSS) ObtainShares(zi, xi []Cipher) error {
	if vss.ABVSSR == nil {
		return errors.New("not a receiver")
	}

	for i := 0; i < vss.batchsize; i++ {
		tmp, err := vss.sigma.Decrypt(vss.sk, zi)
		if err != nil {
			return err
		}
		vss.fshares[i] = tmp
	}
	for i := 0; i < vss.batchsize; i++ {
		tmp, err := vss.sigma.Decrypt(vss.sk, xi)
		if err != nil {
			return err
		}
		vss.gshares[i] = tmp
	}
	return nil
}

func (vss *ABVSS) ConstructLCM(r [][]*big.Int) ([]*big.Int, error) {
	if vss.ABVSSR == nil {
		return nil, errors.New("not a receiver")
	}
	lcm := make([]*big.Int, vss.vnum)
	for i := 0; i < vss.vnum; i++ {
		tmp, err := utils.DotProduct(vss.fshares, r[i])
		if err != nil {
			return nil, err
		}
		lcm[i] = new(big.Int).Add(tmp, vss.gshares[i])
	}
	return lcm, nil
}

func (vss *ABVSS) ShareRecovery() error {
	if vss.ABVSSR == nil {
		return errors.New("not a receiver")
	}
	if !vss.complain {
		return errors.New("not a complain node")
	}
	if len(vss.qlist) < vss.degree+1 {
		return errors.New("invalid Q list")
	}
	xlist := make([]*big.Int, len(vss.qlist))
	ylist := make([]*big.Int, len(vss.qlist))
	for i := 0; i < len(vss.qlist); i++ {
		xlist[i] = new(big.Int).SetInt64(int64(vss.qlist[i].index)
		ylist[i] = vss.qlist[i].fj
	}
	f ,err := polynomial.LagrangeInterpolation(xlist,ylist,vss.p)
	if err != nil {
		return err
	}
	f.EvalMod(vss.index, vss.)
	return nil
}
