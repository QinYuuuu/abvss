package abvss

import (
	"errors"
	"github.com/QinYuuuu/abvss/crypto/elgamal"
	"go.dedis.ch/kyber/v3"
	"log"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
)

func (vss *ABVSS) DistributorInit(pk []kyber.Point, s []*big.Int) error {
	if len(pk) != vss.nodenum {
		return errors.New("node number mismatch PK number")
	}
	if len(s) != vss.batchsize {
		return errors.New("secret number mismatch batchsize")
	}
	vss.ABVSSD = &ABVSSD{
		pk:     pk,
		secret: s,
		polyf:  make([]polynomial.Polynomial, vss.batchsize),
		polyg:  make([]polynomial.Polynomial, vss.vnum),
	}
	return nil
}

func (vss *ABVSS) SamplePoly() error {
	if vss.ABVSSD == nil {
		return errors.New("not a distributor")
	}
	for i := 0; i < vss.batchsize; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.p)
		if err != nil {
			return err
		}
		err = poly.SetCoefficientBig(0, vss.secret[i])
		if err != nil {
			return err
		}
		vss.polyf[i] = poly
	}
	for i := 0; i < vss.vnum; i++ {
		poly, err := polynomial.NewRand(vss.degree, vss.p)
		if err != nil {
			return err
		}
		vss.polyg[i] = poly
	}
	return nil
}

func (vss *ABVSS) GenerateShares(index int) ([]kyber.Point, []kyber.Point, []kyber.Point, []kyber.Point, error) {
	if vss.ABVSSD == nil {
		return nil, nil, nil, nil, errors.New("not a distributor")
	}
	zix := make([]kyber.Point, vss.batchsize)
	ziy := make([]kyber.Point, vss.batchsize)
	xix := make([]kyber.Point, vss.vnum)
	xiy := make([]kyber.Point, vss.vnum)
	for i := 0; i < vss.batchsize; i++ {
		fi := vss.polyf[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmpx, tmpy, r := elgamal.Encrypt(vss.Curve, vss.pk[index], fi.Bytes())
		if len(r) > 0 {
			log.Printf("remainder %s", r)
		}
		zix[i] = tmpx
		ziy[i] = tmpy
	}
	for i := 0; i < vss.vnum; i++ {
		gi := vss.polyg[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmpx, tmpy, r := elgamal.Encrypt(vss.Curve, vss.pk[index], gi.Bytes())
		if len(r) > 0 {
			log.Printf("remainder %s", r)
		}
		xix[i] = tmpx
		xiy[i] = tmpy
	}
	return zix, ziy, xix, xiy, nil
}
