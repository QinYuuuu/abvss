package nodes

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/utils/polynomial"
)

type ABVSS struct {
	degree    int
	counter   int
	p         *big.Int
	randState *rand.Rand
	batchsize int
	vnum      int
	receiver  bool
	verifier  bool
	*ABVSSD
	*ABVSSR
	*ABVSSV
}

type ABVSSD struct {
	secret []*big.Int
	polyf  []polynomial.Polynomial
	polyg  []polynomial.Polynomial
	shares [][]*big.Int
}

type ABVSSR struct {
}

type ABVSSV struct {
	index    []int
	happy1   bool
	happy2   bool
	unhappy1 bool
	unhappy2 bool
	unsure   bool
}

func (vss *ABVSS) Init(s []*big.Int) {
	vss.secret = s
}

func (vss *ABVSS) GenerateShares() error {
	if vss.ABVSSD == nil {
		return errors.New("not a distributor")
	}
	for i := 0; i < vss.batchsize; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.randState, vss.p)
		if err != nil {
			return err
		}
		poly.SetCoefficientBig(0, vss.secret[i])
		if err != nil {
			return err
		}
		vss.polyf[i] = poly
	}
	for i := 0; i < vss.vnum; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.randState, vss.p)
		if err != nil {
			return err
		}
		vss.polyg[i] = poly
	}
	return nil
}

func (vss *ABVSS) ConstructLinearCombinations() {

}
