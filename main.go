package main

import (
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func main() {
	curve := secp256k1.S256()
	curve.Params()
}
