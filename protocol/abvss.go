package protocol

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/crypto/onesidedvoting"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
)

const RAND_SEED = 1

type ABVSS struct {
	index     int
	nodeid    int
	degree    int
	nodenum   int
	p         *big.Int
	randState *rand.Rand
	batchsize int
	vnum      int
	onesidedvoting.OSV

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
		lcm   []*big.Int
	}
	jlist []int
}

func NewVSS(index, nodeid, nodenum, degree, batchsize, vnum int, p *big.Int) (*ABVSS, error) {
	if nodenum < 3*degree+1 {
		return nil, errors.New("must satisfy n >= 3f+1")
	}
	if batchsize <= 0 || vnum <= 0 {
		return nil, errors.New("batchsize/vnum must >= 1")
	}
	return &ABVSS{
		index:     index,
		nodeid:    nodeid,
		degree:    degree,
		nodenum:   nodenum,
		p:         p,
		batchsize: batchsize,
		vnum:      vnum,
		randState: rand.New(rand.NewSource(RAND_SEED)),
	}, nil
}
