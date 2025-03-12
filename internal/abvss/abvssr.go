package abvss

import (
<<<<<<< HEAD
<<<<<<< HEAD
	"errors"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
=======
	"abvss/crypto/elgamal"
	"errors"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/crypto/elgamal"
	"errors"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	"go.dedis.ch/kyber/v3"
	"math/big"
	"math/rand"

<<<<<<< HEAD
<<<<<<< HEAD
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
=======
	"abvss/crypto/utils"
	"abvss/crypto/utils/polynomial"
>>>>>>> 19b0d27 (Initial commit)
=======
	"abvss/crypto/utils"
	"abvss/crypto/utils/polynomial"
<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
)

const ReceiverRandSeed = 10

func (vss *ABVSS) ReceiverInit(sk kyber.Scalar) {
	vss.ABVSSR = &ABVSSR{
		sk:           sk,
		fshares:      make([]*big.Int, vss.batchsize),
		gshares:      make([]*big.Int, vss.vnum),
		xix:          make([][]kyber.Point, vss.nodenum),
		xiy:          make([][]kyber.Point, vss.nodenum),
		zix:          make([][]kyber.Point, vss.nodenum),
		ziy:          make([][]kyber.Point, vss.nodenum),
		Received:     make(chan bool),
		randombeacon: rand.New(rand.NewSource(ReceiverRandSeed)),
	}
}

func (vss *ABVSS) GetShares() ([]*big.Int, []*big.Int) {
	return vss.fshares, vss.gshares
}

func (vss *ABVSS) ObtainShares(zix, ziy, xix, xiy []kyber.Point, index int) error {
	if vss.ABVSSR == nil {
		return errors.New("not a receiver")
	}
	if len(zix) != vss.batchsize || len(ziy) != vss.batchsize {
		return errors.New("insufficient zi")
	}
	if len(xix) != vss.vnum || len(xix) != vss.vnum {
		return errors.New("insufficient xi")
	}
	vss.zix[index] = zix
	vss.ziy[index] = ziy
	vss.xix[index] = xix
	vss.xiy[index] = xiy
	if index == vss.nodeid {
		for i := 0; i < vss.batchsize; i++ {
			tmp, err := elgamal.Decrypt(vss.Curve, vss.sk, zix[i], ziy[i])

			if err != nil {
				/*
					log.Printf("wrong zi %v", zi[i])
					return errors.Join(errors.New("decrypt zi failed"), err)*/
				vss.fshares[i] = utils.RandomNum(vss.p)
			} else {
				vss.fshares[i] = new(big.Int).SetBytes(tmp)
			}
			//vss.fshares[i] = zi[i]
		}
		for i := 0; i < vss.vnum; i++ {

			tmp, err := elgamal.Decrypt(vss.Curve, vss.sk, xix[i], xiy[i])

			if err != nil {
				/*
					log.Printf("wrong xi %v", xi[i])
					return errors.Join(errors.New("decrypt xi failed"), err)*/
				vss.gshares[i] = utils.RandomNum(vss.p)
			} else {
				vss.gshares[i] = new(big.Int).SetBytes(tmp)
			}
			//vss.gshares[i] = xi[i]
		}
		vss.Received <- true
	}
	return nil
}

func (vss *ABVSS) ConstructLCM() ([]*big.Int, error) {
	if vss.ABVSSR == nil {
		return nil, errors.New("not a receiver")
	}
	lcm := make([]*big.Int, vss.vnum)
	r := make([][]*big.Int, vss.vnum)
	for i := 0; i < vss.vnum; i++ {
		r[i] = make([]*big.Int, vss.batchsize)
		for j := 0; j < vss.batchsize; j++ {
			r[i][j] = new(big.Int).Mod(new(big.Int).SetInt64(vss.randombeacon.Int63()), vss.p)
			//fmt.Printf("%v %v %v\n", i, j, r[i][j])
		}
	}
	for i := 0; i < vss.vnum; i++ {
		//fmt.Println(r[i])
		tmp, err := utils.DotProduct(vss.fshares, r[i])
		//fmt.Printf("node %v get fshares %v\n", vss.nodeid, vss.fshares)
		if err != nil {
			return nil, err
		}
		lcm[i] = new(big.Int).Mod(new(big.Int).Add(tmp, vss.gshares[i]), vss.p)
		//fmt.Printf("node %v %v li:%v\n", vss.nodeid, i, lcm[i])
	}
	return lcm, nil
}

func (vss *ABVSS) GetRecoverShares(sk kyber.Scalar, index int, r [][]*big.Int) error {
	fj := make([]*big.Int, vss.batchsize)
	for i := 0; i < vss.batchsize; i++ {
		tmp, err := elgamal.Decrypt(vss.Curve, sk, vss.zix[index][i], vss.ziy[index][i])
		if err != nil {
			return err
		}
		fj[i] = new(big.Int).SetBytes(tmp)
	}
	gj := make([]*big.Int, vss.vnum)
	for i := 0; i < vss.vnum; i++ {
		tmp, err := elgamal.Decrypt(vss.Curve, sk, vss.xix[index][i], vss.xiy[index][i])
		if err != nil {
			return err
		}
		fj[i] = new(big.Int).SetBytes(tmp)
	}
	lcm := make([]*big.Int, vss.vnum)
	for i := 0; i < vss.vnum; i++ {
		tmp, err := utils.DotProduct(fj, r[i])
		if err != nil {
			return err
		}
		lcm[i] = new(big.Int).Add(tmp, gj[i])
	}
	if true {
		vss.qlist[index] = fj
	}
	return nil
}

func (vss *ABVSS) Complain() error {

	return errors.New("do not need to complain")
}

func (vss *ABVSS) HandleComplain(sk kyber.Scalar, index int) error {
<<<<<<< HEAD
<<<<<<< HEAD
	
=======

>>>>>>> 19b0d27 (Initial commit)
=======

<<<<<<< HEAD
>>>>>>> 19b0d27 (Initial commit)
=======
>>>>>>> 19b0d27dd7814a36bd7c868d1a65de42cc91f792
>>>>>>> 777a377d3fc136707c33ac2497d96c7e6cabe03b
	return errors.New("do not need to complain")
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
	ylist := make([][]*big.Int, vss.batchsize)
	for i, index := range vss.jlist {
		xlist[i] = new(big.Int).SetInt64(int64(index))
		ylist[i] = vss.qlist[index]
	}
	for i := 0; i < vss.batchsize; i++ {
		f, err := polynomial.LagrangeInterpolation(xlist, ylist[i], vss.p)
		if err != nil {
			return err
		}
		vss.fshares[i] = f.EvalMod(new(big.Int).SetInt64(int64(vss.nodeid)), vss.p)
	}
	return nil
}
