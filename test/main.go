package main

import (
	"crypto/elliptic"
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/reedsolomon"
	"github.com/QinYuuuu/abvss/crypto/utils/polynomial"
	"github.com/vivint/infectious"
	"math/big"
)

func main() {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	c := elliptic.P224()
	param := c.Params()
	p := param.P
	//s := utils.RandomNum(p)
	poly, _ := polynomial.NewRand(F, p)
	_ = poly.SetCoefficientBig(0, new(big.Int).SetInt64(1))
	_ = poly.SetCoefficientBig(1, new(big.Int).SetInt64(1))
	shares := make([]infectious.Share, N)
	fmt.Println(poly)
	for i := 0; i < F+1; i++ {
		fmt.Println(poly.GetCoefficient(i))
	}
	for i := 0; i < N; i++ {
		fi := poly.EvalMod(new(big.Int).SetInt64(int64(i+1)), p)

		//pad := make([]byte, len(p.Bytes())-len(fi.Bytes()))
		//data := append(pad, fi.Bytes()...)
		shares[i].Data = fi.Bytes()
		shares[i].Number = i
		fmt.Printf("shares %v, %v, %v\n", i, shares[i], len(shares[i].Data))
		fmt.Printf("shares %v, %v\n", i, new(big.Int).SetBytes(shares[i].Data))
	}
	message := append(shares[0].Data, shares[1].Data...)
	rscode1 := reedsolomon.NewRScode(F+1, N)
	rscode2 := reedsolomon.NewReedSolomonCode(F+1, N)
	T1 := rscode1.EncodeNoPadding(message)
	fmt.Println("=====")
	for i := 0; i < N; i++ {
		fmt.Printf("shares %v, %v\n", i, T1[i])
	}
	T2, err := rscode2.Encode(message)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("=====")
	for i := 0; i < N; i++ {
		fmt.Printf("shares %v, %v\n", i, T2[i])
	}

	/*
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
	*/
	/*
		for i := 0; i < N; i++ {
			fmt.Printf("shares %v, %v, %v\n", i, shares[i], len(shares[i].Data))
		}
		message, err := rscode1.Decode(shares)
		if err != nil {
			fmt.Printf("error in rs.decode %v\n", err)
			return
		}
		fmt.Println(message)*/
	/*
		for r := 0; r < F+1; r++ {
			rscode1 := reedsolomon.NewRScode(F+1, N)
			message, err := rscode1.Decode(shares)
			if err != nil {
				fmt.Printf("error in rs.decode %v\n", err)
				return
			}
			fmt.Println(message)
			T := rscode1.Encode(message)
			fmt.Println(T)
			flag := 0
			for i := 0; i < 2*F+r+1; i++ {
				if bytes.Equal(T[i].Data, shares[i].Data) {
					flag++
				}
			}
			fmt.Println("flag ", flag)
			if flag >= 2*F+1 {
				break
			} else {
				continue
			}
		}*/
}
