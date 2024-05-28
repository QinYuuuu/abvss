package erasurecode

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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

	eschunk2 := make([]ErasureCodeChunk, N-F)
	for i := 0; i < N-F; i++ {
		eschunk2[i] = &rschunk[i]
	}

	var message Payload
	err = codec.Decode(eschunk2, &message)
	assert.Nil(t, err, "err in RSDec")
	fmt.Println(string(message.([]byte)))
}

func TestReedSolomonCode_Reconstruct(t *testing.T) {
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
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, rschunk[i].Data)
	}

	eschunk2 := make([]ErasureCodeChunk, N)

	for i := 0; i < N; i++ {
		eschunk2[i] = &rschunk[i]
	}

	eschunk2 = eschunk2[1:]
	for i := 0; i < N-1; i++ {
		fmt.Printf("the %v chunk %v\n", eschunk2[i].Index(), eschunk2[i].GetData())
	}
	tmp, err := rscode.Reconstruct(eschunk2)
	assert.Nil(t, err, "err in RSReconstruct")
	fmt.Println(tmp)
	/*
		var message2 Payload
		err = rscode.Decode(eschunk2, &message2)
		assert.Nil(t, err, "err in RSDec")
		fmt.Println(string(message2.([]byte)))*/
}
