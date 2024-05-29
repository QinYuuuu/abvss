package protocol

import (
	"math/big"
)

type PublicKey interface {
	Encrypt(s *big.Int) (Cipher, error)
}

type SecretKey interface {
	Decrypt(c Cipher) (*big.Int, error)
}

type Cipher interface{}
