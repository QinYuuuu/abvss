package elgamal

import (
	"fmt"
	"go.dedis.ch/kyber/v3/group/curve25519"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"testing"
)

func TestElGamaledwards25519(t *testing.T) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	// Create a public/private keypair
	a := suite.Scalar().Pick(suite.RandomStream()) // Alice's private key
	A := suite.Point().Mul(a, nil)                 // Alice's public key
	fmt.Printf("Maximum number of bytes: %v\n", suite.Point().EmbedLen())
	// ElGamal-encrypt a message using the public key.
	m := []byte("The quick brown fox")
	K, C, _ := Encrypt(suite, A, m)

	// Decrypt it using the corresponding private key.
	mm, err := Decrypt(suite, a, K, C)

	// Make sure it worked!
	if err != nil {
		fmt.Println("decryption failed: " + err.Error())
	}
	if string(mm) != string(m) {
		fmt.Println("decryption produced wrong output: " + string(mm))
		return
	}
	fmt.Println("Decryption succeeded: " + string(mm))

	// Output:
	// Decryption succeeded: The quick brown fox
}

func TestElGamalCurve25519(t *testing.T) {
	suite := curve25519.NewBlakeSHA256Curve25519(true)
	// Create a public/private keypair
	a := suite.Scalar().Pick(suite.RandomStream()) // Alice's private key
	A := suite.Point().Mul(a, nil)                 // Alice's public key
	fmt.Printf("Maximum number of bytes: %v", suite.Point().EmbedLen())
	// ElGamal-encrypt a message using the public key.
	m := []byte("The quick brown fox")
	K, C, _ := Encrypt(suite, A, m)

	// Decrypt it using the corresponding private key.
	mm, err := Decrypt(suite, a, K, C)

	// Make sure it worked!
	if err != nil {
		fmt.Println("decryption failed: " + err.Error())
	}
	if string(mm) != string(m) {
		fmt.Println("decryption produced wrong output: " + string(mm))
		return
	}
	fmt.Println("Decryption succeeded: " + string(mm))

	// Output:
	// Decryption succeeded: The quick brown fox
}

/*
func TestElGamalBLS12381(t *testing.T) {
	suite := bls.NewSuiteBLS12381()
	// Create a public/private keypair
	a := suite.Scalar().Pick(suite.RandomStream()) // Alice's private key
	A := suite.Point().Mul(a, nil)                 // Alice's public key

	// ElGamal-encrypt a message using the public key.
	m := []byte("The quick brown fox")
	K, C, _ := Encrypt(suite, A, m)

	// Decrypt it using the corresponding private key.
	mm, err := Decrypt(suite, a, K, C)

	// Make sure it worked!
	if err != nil {
		fmt.Println("decryption failed: " + err.Error())
	}
	if string(mm) != string(m) {
		fmt.Println("decryption produced wrong output: " + string(mm))
		return
	}
	fmt.Println("Decryption succeeded: " + string(mm))

	// Output:
	// Decryption succeeded: The quick brown fox
}
*/
