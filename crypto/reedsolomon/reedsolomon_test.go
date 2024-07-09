package reedsolomon

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestReedSolomonCode(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message!")
	fmt.Println(data)
	rschunk, err := rscode.Encode(data)
	assert.Nil(t, err, "err in RSEnc")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}
	fmt.Printf("datasize %v\n", rschunk[0].DataSize)

	rschunk2 := make([]ReedSolomonChunk, N-F)
	for i := 0; i < N-F; i++ {
		rschunk2[i] = rschunk[i]
	}

	message, err := rscode.Decode(rschunk2)
	assert.Nil(t, err, "err in RSDec")
	fmt.Println(string(message))
}

// If some chunk missed
func TestReedSolomonCode_Reconstruct1(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")
	fmt.Printf("the message %v\n", data)

	rschunk, err := rscode.Encode(data)
	assert.Nil(t, err, "err in RSEnc")

	fmt.Println("the init chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}

	rschunk2 := make([]ReedSolomonChunk, N-1)
	fmt.Println("missing a random chunk")
	flag := rand.Int() % 4
	i := 0
	for ; i < N-1; i++ {
		if i == flag {
			break
		}
		rschunk2[i] = rschunk[i]
	}
	for ; i < N-1; i++ {
		rschunk2[i] = rschunk[i+1]
	}
	for i := 0; i < N-1; i++ {
		fmt.Printf("the %v chunk %v\n", rschunk2[i].Index(), rschunk2[i].GetData())
	}
	rechunk, err := rscode.Reconstruct(rschunk2)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk[i].GetData())
	}
}

// If some chunk changed
func TestReedSolomonCode_Reconstruct2(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")
	fmt.Printf("the message %v\n", data)

	rschunk, err := rscode.Encode(data)
	assert.Nil(t, err, "err in RSEnc")

	fmt.Println("the init chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}

	fmt.Println("change a random chunk")
	flag := rand.Int() % 4
	rschunk[flag].Data = []byte("wrong")
	rechunk, err := rscode.Reconstruct(rschunk)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk[i].GetData())
	}
}

// If the order of chunks change, RSCode.Decode may return error or return a wrong message
func TestReedSolomonCode_Reconstruct3(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")

	rschunk, err := rscode.Encode(data)
	assert.Nil(t, err, "err in RSEnc")

	fmt.Println("the init chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}

	fmt.Println("random change the order")
	flag1 := rand.Int() % 3
	flag2 := rand.Int() % 3
	for flag1 == flag2 {
		flag2 = rand.Int() % 3
	}
	tmp := rschunk[flag1].Index()
	rschunk[flag1].Idx = rschunk[flag2].Index()
	rschunk[flag2].Idx = tmp
	for i := 0; i < N-1; i++ {
		fmt.Printf("the %v chunk %v\n", rschunk[i].Index(), rschunk[i].GetData())
	}
	rschunk2 := make([]ReedSolomonChunk, N-1)
	for i := 0; i < N-1; i++ {
		rschunk2[i] = rschunk[i]
	}
	rechunk2, err := rscode.Reconstruct(rschunk2)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk2[i].GetData())
	}
	message, err := rscode.Decode(rschunk2)
	assert.Nil(t, err, "err in RSDec")
	fmt.Println(string(message))
}
