package protocol

import (
	"math/big"

	"github.com/QinYuuuu/abvss/crypto/paillier"
)

type PublicKey interface {
	Encrypt(s *big.Int) (Cipher, error)
}

type SecretKey interface {
	Decrypt(c Cipher) (*big.Int, error)
}

type Cipher interface{}

type paillierPubKey struct {
	*paillier.PublicKey
}

type paillierSecretKey struct {
	*paillier.PrivateKey
}

func (pk *paillierPubKey) Encrypt(s *big.Int) (Cipher, error) {
	c, _, err := pk.PublicKey.Encrypt(s)
	return c, err
}

func (sk *paillierSecretKey) Decrypt(c Cipher) (*big.Int, error) {
	var c1 *big.Int
	c1 = c.(*big.Int)
	return sk.PrivateKey.Decrypt(c1)

}
