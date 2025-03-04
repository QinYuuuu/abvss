package erasurecode

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
	var codec ErasureCode = rscode
	data := []byte("a test message")

	var payload Payload = data

	eschunk, err := codec.Encode(payload)
	assert.Nil(t, err, "err in RSEnc")
	rschunk := make([]ReedSolomonChunk, N)
	for i := 0; i < N; i++ {
		if tmp, ok := eschunk[i].(*ReedSolomonChunk); ok {
			rschunk[i] = *tmp
			//fmt.Println("Ok Value =", rschunk, "Ok =", ok)
		} else {
			fmt.Println("Failed Value =", rschunk, "Ok =", ok)
		}
	}
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}
	fmt.Printf("datasize %v\n", rschunk[0].DataSize)
	eschunk2 := make([]ErasureCodeChunk, N-F)
	for i := 0; i < N-F; i++ {
		eschunk2[i] = &rschunk[i]
	}

	var message Payload
	err = codec.Decode(eschunk2, &message)
	assert.Nil(t, err, "err in RSDec")
	fmt.Println(string(message.([]byte)))
}

// If some chunks miss, RSCode.Decode may return error or return a wrong message
func TestReedSolomonCode_Reconstruct1(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")
	fmt.Printf("the message %v\n", data)
	var payload Payload = data

	eschunk, err := rscode.Encode(payload)
	assert.Nil(t, err, "err in RSEnc")
	rschunk := make([]ReedSolomonChunk, N)
	for i := 0; i < N; i++ {
		if tmp, ok := eschunk[i].(*ReedSolomonChunk); ok {
			rschunk[i] = *tmp
			//fmt.Println("Ok Value =", rschunk, "Ok =", ok)
		} else {
			fmt.Println("Failed Value =", rschunk, "Ok =", ok)
		}
	}
	fmt.Println("the init chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}

	eschunk2 := make([]ErasureCodeChunk, N-1)
	fmt.Println("missing a random chunk")
	flag := rand.Int() % 4
	i := 0
	for ; i < N-1; i++ {
		if i == flag {
			break
		}
		eschunk2[i] = &rschunk[i]
	}
	for ; i < N-1; i++ {
		eschunk2[i] = &rschunk[i+1]
	}
	for i := 0; i < N-1; i++ {
		fmt.Printf("the %v chunk %v\n", eschunk2[i].Index(), eschunk2[i].GetData())
	}
	rechunk, err := rscode.Reconstruct(eschunk2)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk[i].GetData())
	}
}

// If the order of chunks change, RSCode.Decode may return error or return a wrong message
func TestReedSolomonCode_Reconstruct2(t *testing.T) {
	N := 4 //"number of servers in the cluster"
	F := 1 //"number of faulty servers to tolerate"
	rscode := NewReedSolomonCode(N-2*F, N)
	data := []byte("a test message")

	var payload Payload = data

	eschunk, err := rscode.Encode(payload)
	assert.Nil(t, err, "err in RSEnc")
	rschunk := make([]ReedSolomonChunk, N)
	for i := 0; i < N; i++ {
		if tmp, ok := eschunk[i].(*ReedSolomonChunk); ok {
			rschunk[i] = *tmp
			//fmt.Println("Ok Value =", rschunk, "Ok =", ok)
		} else {
			fmt.Println("Failed Value =", rschunk, "Ok =", ok)
		}
	}
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
	eschunk2 := make([]ErasureCodeChunk, N-1)
	for i := 0; i < N-1; i++ {
		eschunk2[i] = &rschunk[i]
	}
	rechunk2, err := rscode.Reconstruct(eschunk2)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println("reconstruct chunks")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rechunk2[i].GetData())
	}
	var message Payload
	rscode.Decode(eschunk2, &message)
	fmt.Println(message)
}
