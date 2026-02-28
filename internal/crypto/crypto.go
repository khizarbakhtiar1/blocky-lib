package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/khizar/bc-lib/pkg/types"
)

// Keccak256 computes the Keccak256 hash of the input
func Keccak256(data ...[]byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	for _, b := range data {
		hasher.Write(b)
	}
	return hasher.Sum(nil)
}

// Keccak256Hash computes the Keccak256 hash and returns a Hash type
func Keccak256Hash(data ...[]byte) types.Hash {
	hash := Keccak256(data...)
	h, _ := types.NewHash(hash)
	return h
}

// PubkeyToAddress converts an ECDSA public key to an Ethereum address
func PubkeyToAddress(pubkey *ecdsa.PublicKey) types.Address {
	// Serialize the public key (uncompressed, without the 0x04 prefix)
	// Each coordinate must be exactly 32 bytes, left-padded with zeros
	xBytes := make([]byte, 32)
	yBytes := make([]byte, 32)
	xB := pubkey.X.Bytes()
	yB := pubkey.Y.Bytes()
	copy(xBytes[32-len(xB):], xB)
	copy(yBytes[32-len(yB):], yB)
	pubBytes := append(xBytes, yBytes...)

	// Keccak-256 hash of the public key
	hash := Keccak256(pubBytes)

	// Take the last 20 bytes as the address
	var addr types.Address
	copy(addr[:], hash[12:])
	return addr
}

// GenerateKey generates a new ECDSA private key
func GenerateKey() (*ecdsa.PrivateKey, error) {
	// For now, we'll use a simplified version
	// In production, we should use secp256k1 specifically
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

// Sign signs a hash with a private key
// Returns signature in [R || S || V] format where V is 0 or 1
func Sign(hash []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes")
	}

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}

	// Serialize signature
	sig := make([]byte, 65)
	copy(sig[0:32], r.Bytes())
	copy(sig[32:64], s.Bytes())
	// V will be set based on chain ID during transaction signing
	sig[64] = 0

	return sig, nil
}

// VerifySignature verifies an ECDSA signature against a public key and hash.
// pubkey must be a 64-byte uncompressed public key (X || Y, without the 0x04 prefix).
// signature must be 65 bytes in [R || S || V] format.
func VerifySignature(pubkey, hash, signature []byte) bool {
	if len(signature) != 65 || len(pubkey) != 64 || len(hash) != 32 {
		return false
	}

	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])

	curve := S256()
	x := new(big.Int).SetBytes(pubkey[0:32])
	y := new(big.Int).SetBytes(pubkey[32:64])
	pub := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	return ecdsa.Verify(pub, hash, r, s)
}
