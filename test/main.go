package main

import (
	"bytes"
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/reedsolomn"
	"github.com/QinYuuuu/abvss/crypto/utils"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
	"math/big"
	"math/rand"
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
	length := (F + 1) * len(p.Bytes())
	for i := 0; i < N; i++ {
		fi := poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)
		shares[i].Data = fi.Bytes()
		shares[i].Idx = i
		shares[i].DataSize = length
		fmt.Printf("shares %v, %v, %v\n", i, shares[i], len(fi.Bytes()))
	}
	indexlist := make([]bool, N)
	for i := 0; i < N; i++ {
		indexlist[i] = true
	}

	for i := 0; i < F; i++ {
		for {
			byzantineindex := rand.Int() % N
			if indexlist[byzantineindex] {
				indexlist[byzantineindex] = false
				break
			} else {
				continue
			}
		}
	}
	fmt.Println("the corrupted shares", indexlist)
	for i := 0; i < N; i++ {
		if !indexlist[i] {
			shares[i].Data = utils.RandomNum(p).Bytes()
		}
	}

	for i := 0; i < N; i++ {
		fmt.Printf("shares %v, %v, %v\n", i, shares[i], len(shares[i].Data))
	}
	for r := 0; r < F+1; r++ {
		rscode1 := reedsolomn.NewReedSolomonCode(F+1, 2*F+r+1)
		message, err := rscode1.Decode(shares[:2*F+r+1])
		if err != nil {
			fmt.Printf("error in rs.decode %v\n", err)
		}
		fmt.Println(message)
		T, err := rscode1.Encode(message)
		if err != nil {
			fmt.Printf("error in rs.encode %v\n", err)
		}
		fmt.Println(T)
		flag := 0
		for i := 0; i < 2*F+r+1; i++ {
			if bytes.Equal(T[i].GetData(), shares[i].GetData()) {
				flag++
			}
		}
		fmt.Println("flag ", flag)
		if flag >= 2*F+1 {
			break
		} else {
			continue
		}
	}

}
