package main

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/reedsolomn"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
	"math/big"
)

func main() {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	s := utils.RandomNum(p)
	poly, _ := polynomial.NewRand(F, p)
	_ = poly.SetCoefficientBig(0, s)
	shares := make([]reedsolomn.ReedSolomonChunk, N)
	length := (2*F+1)*len()
	for i := 0; i < N; i++ {
		fi := poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
		shares[i].Data = fi.Bytes()
		shares[i].Idx = i
	}
	fmt.Println("shares\n", shares[:3])
	rscode1 := reedsolomn.NewReedSolomonCode(F+1, 2*F+1)

	message, err := rscode1.Decode(shares[:3])
	fmt.Println(message)
	fmt.Println(err)
}
