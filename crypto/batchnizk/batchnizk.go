package batchnizk

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/curve"
	"github.com/QinYuuuu/abvss/crypto/hasher"
	"github.com/QinYuuuu/abvss/crypto/paillier"
)

type NIZKProof struct {
	tx, ty *big.Int
	e      *big.Int
	r      *big.Int
	q      *big.Int
}

type BatchNIZK struct {
	curve     curve.Curve
	generator *curve.ECPoint
	p         *big.Int
	num       int
	pk        paillier.PublicKey
}

func NewBatchNIZK(curve curve.Curve, g *curve.ECPoint, p *big.Int, num int, pk paillier.PublicKey) *BatchNIZK {
	zk := &BatchNIZK{
		curve:     curve,
		generator: g,
		p:         p,
		num:       num,
		pk:        pk,
	}
	return zk
}

func (zk BatchNIZK) Prove(fij, rij []*big.Int) (*NIZKProof, error) {
	lenth := len(fij)
	if lenth != len(rij) {
		return nil, errors.New("the input length is different")
	}
	g := zk.generator
	p2 := new(big.Int).Sqrt(zk.p)
	u := utils.RandomNum(zk.p)
	s := utils.RandomNum(zk.p)
	tx, ty := zk.curve.ScalarMult(g.X(), g.Y(), u.Bytes())
	pk := zk.pk

	e, err := pk.EncryptWithR(u, s)
	if err != nil {
		return nil, err
	}

	cij := make([]*big.Int, zk.num)
	for i := 0; i < zk.num; i++ {
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, int64(i))
		if err != nil {
			return nil, err
		}
		m := utils.AppendSlices(tx.Bytes(), ty.Bytes(), e.Bytes(), bytesBuffer.Bytes())
		cij[i] = new(big.Int).SetBytes(hasher.SHA256Hasher(m))
	}

	dot, err := utils.DotProduct(fij, cij)
	if err != nil {
		return nil, err
	}
	Rij := new(big.Int).Add(u, dot)

	pow, err := utils.VecPow(rij, cij)
	if err != nil {
		return nil, err
	}
	Qij := new(big.Int).Mod(new(big.Int).Mul(s, pow), p2)
	proof := &NIZKProof{tx: tx, ty: ty, e: e, r: Rij, q: Qij}
	return proof, nil
}

func (zk BatchNIZK) Verify(Aijx, Aijy, zij []*big.Int, pi *NIZKProof) (bool, error) {
	lenth := len(Aijx)
	if lenth != len(zij) || lenth != len(Aijy) || len(zij) != len(Aijy) {
		return false, errors.New("the input length is different")
	}
	p2 := new(big.Int).Sqrt(zk.p)
	cij := make([]*big.Int, zk.num)
	for i := 0; i < zk.num; i++ {
		bytesBuffer := bytes.NewBuffer([]byte{})
		err := binary.Write(bytesBuffer, binary.BigEndian, int64(i))
		if err != nil {
			return false, err
		}
		m := utils.AppendSlices(pi.tx.Bytes(), pi.ty.Bytes(), pi.e.Bytes(), bytesBuffer.Bytes())
		cij[i] = new(big.Int).SetBytes(hasher.SHA256Hasher(m))
	}

	left1x, lef1y := zk.curve.ScalarMult(zk.generator.X(), zk.generator.Y(), pi.r.Bytes())
	dotx, doty, err := curve.DotProductGroup(zk.curve, cij, Aijx, Aijy)
	if err != nil {
		return false, err
	}
	right1x, right1y := zk.curve.Add(pi.tx, pi.ty, dotx, doty)

	t1 := false
	if left1x.Cmp(right1x) == 0 && lef1y.Cmp(right1y) == 0 {
		t1 = true
	}

	pow, err := utils.VecPow(zij, cij)
	if err != nil {
		return false, err
	}
	left2 := new(big.Int).Mod(new(big.Int).Mul(pi.e, pow), p2)
	right2, err := zk.pk.EncryptWithR(pi.r, pi.q)
	if err != nil {
		return false, err
	}
	t2 := false
	if left2.Cmp(right2) == 0 {
		t2 = true
	}
	return t1 && t2, nil
}
