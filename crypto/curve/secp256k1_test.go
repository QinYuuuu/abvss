package curve

import (
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"testing"
)

func TestCurveLoad(t *testing.T) {
	var curve Curve
	curve = secp256k1.S256()
	curve.Params()
}
