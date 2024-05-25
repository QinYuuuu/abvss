package protocol

import (
	"math/big"
	"math/rand"
)

type ABDKG struct {
	abvss     []*ABVSS
	index     int
	degree    int
	nodenum   int
	p         *big.Int
	randState *rand.Rand
	batchsize int
	vnum      int
}

func NewDKG(index, nodenum, degree, batchsize, vnum int, p *big.Int) (*ABDKG, error) {
	return &ABDKG{
		index:     index,
		degree:    degree,
		nodenum:   nodenum,
		p:         p,
		batchsize: batchsize,
		vnum:      vnum,
	}, nil
}

func (dkg *ABDKG) SecretSharing() {
	return
}
