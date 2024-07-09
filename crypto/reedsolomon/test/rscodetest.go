package main

import (
	"fmt"
	"github.com/QinYuuuu/abvss/crypto/reedsolomon"
	"math/rand"
)

func main() {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := reedsolomon.NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")
	fmt.Printf("the message %v\n", data)

	rschunk, err := rscode.Encode(data)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("the init chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i])
	}

	fmt.Println("change a random chunk")
	flag := rand.Int() % 4
	rschunk[flag].Data = []byte("!wrong!")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i])
	}
	verify, err := rscode.Verify(rschunk)
	if err != nil {
		fmt.Println("error in Verify:", err)
	}
	fmt.Println("verify chunks", verify)
	rechunk, err := rscode.Reconstruct(rschunk)
	if err != nil {
		fmt.Println("error in reconstruct:", err)
	}
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk[i].GetData())
	}
}
