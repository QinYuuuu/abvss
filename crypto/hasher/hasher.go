// Hasher is a package that provides different kind of hash functions
// like SHA256, SHA512, MD5, Keccak256, Poseidon and MIMC7.
package hasher

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
)

// SHA256Hasher returns the SHA256 hash of given input
//
// It takes one slice of bytes.
//
// Parameters:
// - data: Input data as slice of bytes
//
// Returns:
// - Hashed data as slice of bytes
//
// Example:
//
//	hash := hasher.SHA256Hasher([]byte("1"))
//	fmt.Println(hash)
func SHA256Hasher(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// SHA512Hasher returns the SHA512 hash of given input
//
// It takes one slice of bytes.
//
// Parameters:
// - data: Input data as slice of bytes
//
// Returns:
// - Hashed data as slice of bytes
//
// Example:
//
//	hash := hasher.SHA512Hasher([]byte("1"))
//	fmt.Println(hash)
func SHA512Hasher(data []byte) []byte {
	hash := sha512.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// MD5Hasher returns the MD5 hash of given input
//
// It takes one slice of bytes.
//
// Parameters:
// - data: Input data as slice of bytes
//
// Returns:
// - Hashed data as slice of bytes
//
// Example:
//
//	hash := hasher.MD5Hasher([]byte("1"))
//	fmt.Println(hash)
func MD5Hasher(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}
