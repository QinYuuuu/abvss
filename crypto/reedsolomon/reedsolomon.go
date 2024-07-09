package reedsolomon

import (
	"bytes"
	"fmt"
	"github.com/klauspost/reedsolomon"
	"log"
)

type RSCode interface {
	NewRSCode(d, p int) *RSCode
}

type ReedSolomonCode struct {
	d int // number of data shards
	p int // number of data + check shards
	reedsolomon.Encoder
}

func NewReedSolomonCode(d, p int) *ReedSolomonCode {
	enc, err := reedsolomon.New(d, p-d)
	if err != nil {
		log.Fatalln("error creating RS encoder:", err)
	}
	c := &ReedSolomonCode{
		d:       d,
		p:       p,
		Encoder: enc,
	}
	return c
}

func (rscode *ReedSolomonCode) Encode(input []byte) ([]ReedSolomonChunk, error) {
	output := make([]ReedSolomonChunk, rscode.p)
	datasize := len(input)
	shards, err := rscode.Split(input)
	if err != nil {
		return output, err
	}
	err = rscode.Encoder.Encode(shards)
	if err != nil {
		return output, err
	}
	if len(shards) != rscode.p {
		panic("wrong number of shards")
	}

	for i := 0; i < rscode.p; i++ {
		output[i] = ReedSolomonChunk{
			DataSize: datasize,
			Idx:      i,
			Data:     shards[i],
		}
	}
	return output, nil
}

func (rscode *ReedSolomonCode) Reconstruct(shards []ReedSolomonChunk) ([]ReedSolomonChunk, error) {
	input := make([][]byte, rscode.p)
	for i := 0; i < len(shards); i++ {
		input[shards[i].Index()] = shards[i].GetData()
	}
	err := rscode.Encoder.Reconstruct(input)
	if err != nil {
		return nil, err
	}
	out := make([]ReedSolomonChunk, rscode.p)
	for i, v := range input {
		out[i] = ReedSolomonChunk{
			DataSize: len(v),
			Idx:      i,
			Data:     v,
		}
	}
	return out, err
}

func (rscode *ReedSolomonCode) Verify(shards []ReedSolomonChunk) (bool, error) {
	fmt.Println(shards)
	fmt.Println(shards[0].Size())
	input := make([][]byte, rscode.p)
	for i := 0; i < rscode.p; i++ {
		input[i] = make([]byte, shards[0].Size())
	}

	for _, v := range shards {
		ptr := v
		input[ptr.Idx] = ptr.Data
	}

	fmt.Println(input)
	return rscode.Encoder.Verify(input)
}

func (rscode *ReedSolomonCode) Decode(shards []ReedSolomonChunk) ([]byte, error) {
	// TODO: we are trusting the first shard
	datasize := shards[0].DataSize

	input := make([][]byte, rscode.p)
	for _, v := range shards {
		ptr := v
		input[ptr.Idx] = ptr.Data
	}
	err := rscode.Encoder.Reconstruct(input)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = rscode.Encoder.Join(buf, input, datasize)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
