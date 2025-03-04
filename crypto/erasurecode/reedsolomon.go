package erasurecode

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	//"math"
	//"crypto/sha256"
	"github.com/klauspost/reedsolomon"
)

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

type ReedSolomonChunk struct {
	DataSize int
	Idx      int
	Data     []byte
	//Merkle   []byte
}

func (c *ReedSolomonChunk) Index() int {
	return c.Idx
}

func (c *ReedSolomonChunk) GetData() []byte {
	return c.Data
}

func (c *ReedSolomonChunk) Size() int {
	return len(c.Data)
}

func (rscode *ReedSolomonCode) Encode(input Payload) ([]ErasureCodeChunk, error) {
	output := make([]ErasureCodeChunk, rscode.p)
	buf := &bytes.Buffer{}
	encoder := gob.NewEncoder(buf)
	// this is tricky. why indirect input? it is because if we pass input to gob, it still appears
	// to gob as the concrete type. the receiving end is expecting an interface, and will complain.
	// we use indirect here so that gob cannot figure out the concrete type, and will thus happily
	// encode it as an interface
	err := encoder.Encode(&input)
	if err != nil {
		return output, err
	}

	b := buf.Bytes()
	fmt.Println(string(b))
	datasize := len(b)
	shards, err := rscode.Split(b)
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
		output[i] = &ReedSolomonChunk{
			DataSize: datasize,
			Idx:      i,
			Data:     shards[i],
			//Merkle:   bytes.Repeat([]byte("a"), int(math.Log2(float64(f.p)))*32),
		}
	}
	return output, nil
}

func (rscode *ReedSolomonCode) Reconstruct(shards []ErasureCodeChunk) ([]ErasureCodeChunk, error) {
	input := make([][]byte, rscode.p)
	for i := 0; i < len(shards); i++ {
		input[shards[i].Index()] = shards[i].GetData()
	}
	err := rscode.Encoder.Reconstruct(input)
	if err != nil {
		return nil, err
	}
	out := make([]ErasureCodeChunk, rscode.p)
	for i, v := range input {
		out[i] = &ReedSolomonChunk{
			DataSize: len(v),
			Idx:      i,
			Data:     v,
		}
	}
	return out, err
}

/*
	func (rscode *ReedSolomonCode) Verify(shards []ErasureCodeChunk) (bool, error) {
		fmt.Println(shards)
		fmt.Println(shards[0].Size())
		input := make([][]byte, rscode.p)
		for i := 0; i < rscode.p; i++ {
			input[i] = make([]byte, shards[0].Size())
		}

		for _, v := range shards {
			ptr := v.(*ReedSolomonChunk)
			input[ptr.Idx] = ptr.Data
		}

		fmt.Println(input)
		return rscode.Encoder.Verify(input)
	}
*/
func (rscode *ReedSolomonCode) Decode(shards []ErasureCodeChunk, v *Payload) error {
	// TODO: we are trusting the first shard
	datasize := shards[0].(*ReedSolomonChunk).DataSize

	input := make([][]byte, rscode.p)
	for _, v := range shards {
		ptr := v.(*ReedSolomonChunk)
		input[ptr.Idx] = ptr.Data
	}
	err := rscode.Encoder.Reconstruct(input)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	decoder := gob.NewDecoder(buf)
	err = rscode.Encoder.Join(buf, input, datasize)
	if err != nil {
		return err
	}
	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	return nil
}
