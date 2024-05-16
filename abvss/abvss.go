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

	receiver bool
	verifier bool
	*ABVSSD
	*ABVSSR
	*ABVSSV
}

type ABVSSD struct {
	secret *big.Int
	poly   *polynomial.Polynomial
	shares []*big.Int
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

func (vss *ABVSS) Init() {

}

func (vss *ABVSS) GenerateShares(s *big.Int) error {
	if vss.ABVSSD == nil {
		return errors.New("not a distributor")
	}
	vss.secret = s
	poly, err := polynomial.NewRand(vss.degree, vss.randState, vss.p)
	if err != nil {
		return err
	}
	vss.poly = &poly

	return nil
}
