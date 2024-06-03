package abvss

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/paillier"
	"math/big"
	"math/rand"

	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
)

const ReceiverRandSeed = 10

func (vss *ABVSS) ReceiverInit(sk paillier.PrivateKey) {
	vss.ABVSSR = &ABVSSR{
		sk:           sk,
		fshares:      make([]*big.Int, vss.batchsize),
		gshares:      make([]*big.Int, vss.vnum),
		xi:           make([][]*big.Int, vss.nodenum),
		zi:           make([][]*big.Int, vss.nodenum),
		Received:     false,
		randombeacon: rand.New(rand.NewSource(ReceiverRandSeed)),
	}
}

func (vss *ABVSS) ObtainShares(zi, xi []*big.Int, index int) error {
	if vss.ABVSSR == nil {
		return errors.New("not a receiver")
	}
	if len(zi) != vss.batchsize {
		return errors.New("insufficient zi")
	}
	if len(xi) != vss.vnum {
		return errors.New("insufficient xi")
	}
	vss.zi[index] = zi
	vss.xi[index] = xi
	if index == vss.index {
		for i := 0; i < vss.batchsize; i++ {
			tmp, err := vss.sk.Decrypt(zi[i])

			if err != nil {
				/*
					log.Printf("wrong zi %v", zi[i])
					return errors.Join(errors.New("decrypt zi failed"), err)*/
				vss.fshares[i] = utils.RandomNum(vss.p)
			} else {
				vss.fshares[i] = tmp
			}
			//vss.fshares[i] = zi[i]
		}
		for i := 0; i < vss.vnum; i++ {

			tmp, err := vss.sk.Decrypt(xi[i])

			if err != nil {
				/*
					log.Printf("wrong xi %v", xi[i])
					return errors.Join(errors.New("decrypt xi failed"), err)*/
				vss.gshares[i] = utils.RandomNum(vss.p)
			} else {
				vss.gshares[i] = tmp
			}
			//vss.gshares[i] = xi[i]
		}
		vss.Received = true
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

func (vss *ABVSS) GetRecoverShares(sk services.SecretKey, index int, r [][]*big.Int) error {
	fj := make([]*big.Int, vss.batchsize)
	for i := 0; i < vss.batchsize; i++ {
		tmp, err := sk.Decrypt(vss.zi[index])
		if err != nil {
			return err
		}
		fj[i] = tmp
	}
	gj := make([]*big.Int, vss.vnum)
	for i := 0; i < vss.vnum; i++ {
		tmp, err := sk.Decrypt(vss.xi[index])
		if err != nil {
			return err
		}
		fj[i] = tmp
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

func (vss *ABVSS) HandleComplain() error {
	return nil
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
		vss.fshares[i] = f.EvalMod(new(big.Int).SetInt64(int64(vss.index)), vss.p)
	}
	return nil
}
