package rsa

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// SaveRSAKeyBytesToFile 将RSA密钥字节数组保存到文件
func SaveRSAKeyBytesToFile(privateKeyBytes []byte, publicKeyBytes []byte, privateKeyPath, publicKeyPath string) error {
	// 保存私钥到文件
	if err := os.WriteFile(privateKeyPath, privateKeyBytes, 0600); err != nil {
		return fmt.Errorf("保存私钥文件失败: %w", err)
	}

	// 保存公钥到文件
	if err := os.WriteFile(publicKeyPath, publicKeyBytes, 0644); err != nil {
		return fmt.Errorf("保存公钥文件失败: %w", err)
	}

	return nil
}

// ReadPrivateKeyFromFile 从文件中读取RSA私钥
func ReadPrivateKeyFromFile(privateKeyPath string) ([]byte, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥文件失败: %w", err)
	}

	// 验证密钥是否有效
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid RSA private key PEM data")
	}

	// 尝试解析私钥，确保格式正确
	_, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
	}

	return privateKeyBytes, nil
}

// ReadPublicKeyFromFile 从文件中读取RSA公钥
func ReadPublicKeyFromFile(publicKeyPath string) ([]byte, error) {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("读取公钥文件失败: %w", err)
	}

	// 验证密钥是否有效
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("无效的RSA公钥PEM数据")
	}

	// 尝试解析公钥，确保格式正确
	_, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA公钥失败: %w", err)
	}

	return publicKeyBytes, nil
}

// ParsePrivateKeyFromBytes 从PEM编码的字节数组解析RSA私钥
func ParsePrivateKeyFromBytes(privateKeyBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid RSA private key PEM data")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA私钥失败: %w", err)
	}

	return privateKey, nil
}

// ParsePublicKeyFromBytes 从PEM编码的字节数组解析RSA公钥
func ParsePublicKeyFromBytes(publicKeyBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("invalid RSA public key PEM data")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("解析RSA公钥失败: %w", err)
	}

	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("不是RSA公钥")
	}

	return publicKey, nil
}