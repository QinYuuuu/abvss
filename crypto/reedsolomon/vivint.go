package reedsolomon

import (
	"github.com/vivint/infectious"
)

// RscodeVivint is a reedsolomon coder import from vivint
type RscodeVivint struct {
	fec *infectious.FEC
}

// NewRScode returns a RScoder object
func NewRScode(requried, total int) *RscodeVivint {
	temp, _ := infectious.NewFEC(requried, total)
	coder := &RscodeVivint{
		fec: temp,
	}
	return coder
}

// Encode returns shares of the encoded message
func (coder *RscodeVivint) Encode(msg []byte) []infectious.Share {
	shares := make([]infectious.Share, coder.fec.Total())
	output := func(s infectious.Share) {
		shares[s.Number] = s.DeepCopy() // the memory in s gets reused, so we need to make a deep copy
	}
	paddingMessage := coder.Padding(msg)
	err := coder.fec.Encode(paddingMessage, output)
	if err != nil {
		panic(err)
	}
	return shares
}

// EncodeNoPadding returns shares of the encoded message
func (coder *RscodeVivint) EncodeNoPadding(msg []byte) []infectious.Share {
	shares := make([]infectious.Share, coder.fec.Total())
	output := func(s infectious.Share) {
		shares[s.Number] = s.DeepCopy() // the memory in s gets reused, so we need to make a deep copy
	}
	err := coder.fec.Encode(msg, output)
	if err != nil {
		panic(err)
	}
	return shares
}

func (coder *RscodeVivint) Padding(msg []byte) []byte {
	paddingLength := coder.fec.Required() - (len(msg) % coder.fec.Required())
	paddingMessage := make([]byte, len(msg)+paddingLength)
	copy(paddingMessage, msg)
	paddingMessage[len(paddingMessage)-1] = byte(paddingLength) //p.F+1 == coder.Required() == paddingLength < 256, so byte is enough
	return paddingMessage
}

// Decode returns the original message of the shares
func (coder *RscodeVivint) Decode(shares []infectious.Share) ([]byte, error) {
	result, err := coder.fec.Decode(nil, shares)
	if err != nil {
		return nil, err
	}
	return result[:len(result)-int(result[len(result)-1])], nil
}
