package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/sha3"
)

// Address represents a blockchain address (20 bytes)
type Address [20]byte

// String returns the hex string representation of the address with 0x prefix
// Uses EIP-55 checksum encoding
func (a Address) String() string {
	return toChecksumAddress(a)
}

// StringLower returns the lowercase hex string without checksum
func (a Address) StringLower() string {
	return "0x" + hex.EncodeToString(a[:])
}

// toChecksumAddress returns the EIP-55 checksummed address
func toChecksumAddress(addr Address) string {
	hexAddr := hex.EncodeToString(addr[:])

	// Keccak256 hash of the lowercase hex address
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(hexAddr))
	hash := hasher.Sum(nil)

	result := make([]byte, 42)
	result[0] = '0'
	result[1] = 'x'

	for i, c := range hexAddr {
		if c >= '0' && c <= '9' {
			result[i+2] = byte(c)
		} else {
			// Check if we should uppercase: if the corresponding nibble in hash >= 8
			hashNibble := hash[i/2]
			if i%2 == 0 {
				hashNibble = hashNibble >> 4
			} else {
				hashNibble = hashNibble & 0x0f
			}

			if hashNibble >= 8 {
				result[i+2] = byte(c - 32) // uppercase
			} else {
				result[i+2] = byte(c)
			}
		}
	}

	return string(result)
}

// Bytes returns the address as a byte slice
func (a Address) Bytes() []byte {
	return a[:]
}

// Hex returns the hex string without 0x prefix
func (a Address) Hex() string {
	return hex.EncodeToString(a[:])
}

// IsZero returns true if the address is the zero address
func (a Address) IsZero() bool {
	return a == Address{}
}

// NewAddress creates an Address from a byte slice
func NewAddress(b []byte) (Address, error) {
	if len(b) != 20 {
		return Address{}, fmt.Errorf("invalid address length: expected 20, got %d", len(b))
	}
	var addr Address
	copy(addr[:], b)
	return addr, nil
}

// AddressFromHex creates an Address from a hex string
// Accepts both "0x" prefixed and non-prefixed strings
func AddressFromHex(s string) (Address, error) {
	s = strings.TrimPrefix(s, "0x")
	if len(s) != 40 {
		return Address{}, fmt.Errorf("invalid address hex length: expected 40, got %d", len(s))
	}

	decoded, err := hex.DecodeString(s)
	if err != nil {
		return Address{}, fmt.Errorf("invalid hex string: %w", err)
	}

	return NewAddress(decoded)
}

// MustAddressFromHex creates an Address from a hex string, panics on error
// Useful for constants and test fixtures
func MustAddressFromHex(s string) Address {
	addr, err := AddressFromHex(s)
	if err != nil {
		panic(err)
	}
	return addr
}

// ZeroAddress returns the zero address (0x0000...0000)
func ZeroAddress() Address {
	return Address{}
}

// Common known addresses
var (
	// AddressZero is the zero address
	AddressZero = ZeroAddress()
)

// Validate checks if the address is valid (non-zero)
func (a Address) Validate() error {
	if a.IsZero() {
		return errors.New("address cannot be zero")
	}
	return nil
}

// ValidateChecksum validates that a hex address has the correct EIP-55 checksum
func ValidateChecksum(hexAddr string) bool {
	if !strings.HasPrefix(hexAddr, "0x") {
		return false
	}

	addr, err := AddressFromHex(hexAddr)
	if err != nil {
		return false
	}

	return toChecksumAddress(addr) == hexAddr
}

// Compare compares two addresses
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func (a Address) Compare(b Address) int {
	for i := 0; i < 20; i++ {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return 1
		}
	}
	return 0
}

// Equal checks if two addresses are equal
func (a Address) Equal(b Address) bool {
	return a == b
}
