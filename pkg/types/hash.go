package types

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// Hash represents a 32-byte hash (transaction hash, block hash, etc.)
type Hash [32]byte

// String returns the hex string representation with 0x prefix
func (h Hash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

// Bytes returns the hash as a byte slice
func (h Hash) Bytes() []byte {
	return h[:]
}

// Hex returns the hex string without 0x prefix
func (h Hash) Hex() string {
	return hex.EncodeToString(h[:])
}

// IsZero returns true if the hash is zero
func (h Hash) IsZero() bool {
	return h == Hash{}
}

// NewHash creates a Hash from a byte slice
func NewHash(b []byte) (Hash, error) {
	if len(b) != 32 {
		return Hash{}, fmt.Errorf("invalid hash length: expected 32, got %d", len(b))
	}
	var hash Hash
	copy(hash[:], b)
	return hash, nil
}

// HashFromHex creates a Hash from a hex string
// Accepts both "0x" prefixed and non-prefixed strings
func HashFromHex(s string) (Hash, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s) != 64 {
		return Hash{}, fmt.Errorf("invalid hash hex length: expected 64, got %d", len(s))
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return Hash{}, fmt.Errorf("invalid hex string: %w", err)
	}

	return NewHash(decoded)
}

// MustHashFromHex creates a Hash from a hex string, panics on error
func MustHashFromHex(s string) Hash {
	hash, err := HashFromHex(s)
	if err != nil {
		panic(err)
	}
	return hash
}

// ZeroHash returns the zero hash
func ZeroHash() Hash {
	return Hash{}
}

// Common known hashes
var (
	// HashZero is the zero hash
	HashZero = ZeroHash()
)
