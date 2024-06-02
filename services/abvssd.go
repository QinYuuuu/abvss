package services

import (
	"errors"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
)

func (vss *ABVSS) DistributorInit(pk []PublicKey, s []*big.Int) error {
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
		poly, err := polynomial.NewRand(vss.degree, vss.randState, vss.p)
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
		poly, err := polynomial.NewRand(vss.degree, vss.randState, vss.p)
		if err != nil {
			return err
		}
		vss.polyg[i] = poly
	}
	return nil
}

func (vss *ABVSS) GenerateShares(index int) ([]Cipher, []Cipher, error) {
	if vss.ABVSSD == nil {
		return nil, nil, errors.New("not a distributor")
	}
	zi := make([]Cipher, vss.batchsize)
	xi := make([]Cipher, vss.vnum)
	for i := 0; i < vss.batchsize; i++ {
		fi := vss.polyf[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmp, err := vss.pk[index].Encrypt(fi)
		if err != nil {
			return nil, nil, err
		}
		zi[i] = tmp
	}
	for i := 0; i < vss.vnum; i++ {
		gi := vss.polyg[i].EvalMod(new(big.Int).SetInt64(int64(index+1)), vss.p)
		tmp, err := vss.pk[index].Encrypt(gi)
		if err != nil {
			return nil, nil, err
		}
		xi[i] = tmp
	}
	return zi, xi, nil
}
