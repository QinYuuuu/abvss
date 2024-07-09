package reedsolomon

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

func TestEncode(t *testing.T) {
	N := 4
	F := 1
	coder := NewRScode(F+1, N)

	msg := []byte("a test message!")
	shares := coder.Encode(msg)
	fmt.Println(shares)
	rmsg, _ := coder.Decode(shares)
	if !bytes.Equal(msg, rmsg) {
		t.Error()
	}
}

func TestDecode(t *testing.T) {
	N := 4
	F := 1
	coder := NewRScode(F+1, N)

	msg := []byte("a test message!")
	shares := coder.Encode(msg)
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, shares[i])
	}
	fmt.Println("change a random chunk")
	flag := rand.Int() % 4
	shares[flag].Data = []byte("!wrong!!")
	for i := 0; i < N; i++ {
		fmt.Printf("the %v chunk %v\n", i, shares[i])
	}
	rmsg, _ := coder.Decode(shares)
	if !bytes.Equal(msg, rmsg) {
		t.Error()
	}
}
