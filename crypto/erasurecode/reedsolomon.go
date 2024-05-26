package erasurecode

import (
	"bytes"
	"encoding/gob"
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

func (c *ReedSolomonChunk) Index() int{
	return c.Idx
}

func (c *ReedSolomonChunk) GetData() []byte{
	return c.Data
}

func (c *ReedSolomonChunk) Size() int {
	return len(c.Data)
}

func (f *ReedSolomonCode) Encode(input Payload) ([]ErasureCodeChunk, error) {
	output := make([]ErasureCodeChunk, f.p)
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
	datasize := len(b)
	shards, err := f.Split(b)
	if err != nil {
		return output, err
	}
	err = f.Encoder.Encode(shards)
	if err != nil {
		return output, err
	}
	if len(shards) != f.p {
		panic("wrong number of shards")
	}

	for i := 0; i < f.p; i++ {
		output[i] = &ReedSolomonChunk{
			DataSize: datasize,
			Idx:      i,
			Data:     shards[i],
			//Merkle:   bytes.Repeat([]byte("a"), int(math.Log2(float64(f.p)))*32),
		}
	}
	return output, nil
}

func (f *ReedSolomonCode) Decode(shards []ErasureCodeChunk, v *Payload) error {
	// TODO: we are trusting the first shard
	datasize := shards[0].(*ReedSolomonChunk).DataSize

	input := make([][]byte, f.p)
	for _, v := range shards {
		ptr := v.(*ReedSolomonChunk)
		input[ptr.Idx] = ptr.Data
	}

	err := f.Encoder.Reconstruct(input)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	decoder := gob.NewDecoder(buf)
	err = f.Encoder.Join(buf, input, datasize)
	if err != nil {
		return err
	}
	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	return nil
}