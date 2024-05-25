package protocol

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
)

type ABVSS struct {
	index     int
	degree    int
	nodenum   int
	p         *big.Int
	randState *rand.Rand
	batchsize int
	vnum      int

	sigma Sigma
	*ABVSSD
	*ABVSSR
	*ABVSSV
}

type ABVSSD struct {
	pk     []PublicKey
	secret []*big.Int
	polyf  []polynomial.Polynomial
	polyg  []polynomial.Polynomial
	//shares [][]*big.Int
}

type ABVSSR struct {
	sk       SecretKey
	zi       map[int][]Cipher
	xi       map[int][]Cipher
	fshares  []*big.Int
	gshares  []*big.Int
	complain bool
	qlist    map[int][]*big.Int
}

type ABVSSV struct {
	ilist []struct {
		index int
		lcm   *big.Int
	}
	jlist []int
}

func NewVSS(index, nodenum, degree, batchsize, vnum int, p *big.Int, sigma Sigma) (*ABVSS, error) {
	if nodenum < 3*degree+1 {
		return nil, errors.New("must satisfy n >= 3f+1")
	}
	if batchsize <= 0 || vnum <= 0 {
		return nil, errors.New("batchsize/vnum must >= 1")
	}
	return &ABVSS{
		index:     index,
		degree:    degree,
		nodenum:   nodenum,
		p:         p,
		batchsize: batchsize,
		vnum:      vnum,
		sigma:     sigma,
	}, nil
}
