package abvss

import "math/big"

type PublicKey interface{}

type SecretKey interface{}

type Cipher interface{}

type Sigma interface {
	Encrypt(pk PublicKey, s []*big.Int) Cipher
	Decrypt(sk []SecretKey, c Cipher) []*big.Int
}
