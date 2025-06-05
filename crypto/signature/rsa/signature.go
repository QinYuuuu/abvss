package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// KeyGenBytes return priKey, pubKey, error
func KeyGenBytes(bits int) ([]byte, []byte, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	// 将私钥转换为PKCS#1 DER格式
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	}

	// 编码为PEM格式
	privateKeyBytes := pem.EncodeToMemory(&privateKeyBlock)
	if privateKeyBytes == nil {
		return nil, nil, fmt.Errorf("编码私钥为PEM格式失败")
	}

	// 提取公钥
	publicKey := &privateKey.PublicKey

	// 将公钥转换为PKIX DER格式
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("encode public key failed: %w", err)
	}

	publicKeyBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}

	// 编码为PEM格式
	publicKeyBytes := pem.EncodeToMemory(&publicKeyBlock)
	if publicKeyBytes == nil {
		return nil, nil, fmt.Errorf("encode public key to PEM failed")
	}

	return privateKeyBytes, publicKeyBytes, nil
}

func KeyGen(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	// 提取公钥
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

func Sign(data []byte, privKeyBytes []byte) ([]byte, error) {
	privKey, err := ParsePrivateKeyFromBytes(privKeyBytes)
	if err != nil {
		return nil, err
	}
	hashed := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func Verify(data, signature []byte, publicKeyBytes []byte) error {
	pubKey, err := ParsePublicKeyFromBytes(publicKeyBytes)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256(data)
	err = rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], signature)
	return err
}
