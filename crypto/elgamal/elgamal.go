package elgamal

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/curve25519"
	"go.dedis.ch/kyber/v3/util/random"
)

func Encrypt(group kyber.Group, pubkey kyber.Point, message []byte) (K, C kyber.Point, remainder []byte) {

	// Embed the message (or as much of it as will fit) into a curve point.
	M := group.Point().Embed(message, random.New())
	max := group.Point().EmbedLen()
	if max > len(message) {
		max = len(message)
	}
	remainder = message[max:]
	// ElGamal-encrypt the point to produce ciphertext (K,C).
	k := group.Scalar().Pick(random.New()) // ephemeral private key
	K = group.Point().Mul(k, nil)          // ephemeral DH public key
	S := group.Point().Mul(k, pubkey)      // ephemeral DH shared secret
	C = S.Add(S, M)                        // message blinded with secret
	return
}

func Decrypt(group kyber.Group, prikey kyber.Scalar, K, C kyber.Point) (message []byte, err error) {

	// ElGamal-decrypt the ciphertext (K,C) to reproduce the message.
	S := group.Point().Mul(prikey, K) // regenerate shared secret
	M := group.Point().Sub(C, S)      // use to un-blind the message
	message, err = M.Data()           // extract the embedded data
	return
}

func KeyGenCurve25519(suit *curve25519.SuiteCurve25519) (kyber.Point, kyber.Scalar) {
	a := suit.Scalar().Pick(suit.RandomStream()) // Alice's private key
	A := suit.Point().Mul(a, nil)                // Alice's public key
	return A, a
}
