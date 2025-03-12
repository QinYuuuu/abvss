package abdkg

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/internal/abvss"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/crypto/curve"
=======
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
	"abvss/internal/abvss"
	"math/big"
	"math/rand"

	"abvss/crypto/curve"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
)

type ABDKG struct {
	abvss     []*abvss.ABVSS
	index     int
	degree    int
	nodenum   int
	curve     curve.Curve
	p         *big.Int
	randState *rand.Rand
	batchsize int
	vnum      int
	fshares   [][]*big.Int
	gshares   [][]*big.Int
	Dlist     []int
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

func (dkg *ABDKG) PKComponentsAndProofs() {
	for _, i := range dkg.Dlist {
		Aijx := make([]*big.Int, dkg.batchsize)
		Aijy := make([]*big.Int, dkg.batchsize)
		Bijx := make([]*big.Int, dkg.vnum)
		Bijy := make([]*big.Int, dkg.vnum)
		fij := dkg.fshares[i-1]
		gij := dkg.gshares[i-1]
		for i, fijl := range fij {
			Aijx[i], Aijy[i] = dkg.curve.ScalarBaseMult(fijl.Bytes())
		}
		for i, gijl := range gij {
			Bijx[i], Bijy[i] = dkg.curve.ScalarBaseMult(gijl.Bytes())
		}
		//pij := dkg.zk.Prove(fij, )
	}
}
