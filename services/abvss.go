package services

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"math/big"
	"math/rand"

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

	*ABVSSD
	*ABVSSR
	*ABVSSV
}

func (vss *ABVSS) GetNodeID() int {
	return vss.nodeid
}
func (vss *ABVSS) GetInstanceID() int {
	return vss.index
}

type ABVSSD struct {
	pk     []paillier.PublicKey
	secret []*big.Int
	polyf  []polynomial.Polynomial
	polyg  []polynomial.Polynomial
	//shares [][]*big.Int
}

type ABVSSR struct {
	sk paillier.PrivateKey

	//zi           [][]Cipher
	//xi           [][]Cipher
	zi           [][]*big.Int
	xi           [][]*big.Int
	fshares      []*big.Int
	gshares      []*big.Int
	randombeacon *rand.Rand
	r            [][]*big.Int
	received     bool
	complain     bool
	qlist        map[int][]*big.Int
}

type ABVSSV struct {
	count int
	ilist []struct {
		index int
		lcm   []*big.Int
	}
	jlist []int
	done  bool
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
