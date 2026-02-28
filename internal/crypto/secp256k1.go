package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// S256 returns the secp256k1 curve.
// NOTE: Uses P256 as a stand-in until a proper secp256k1 library
// (e.g. github.com/btcsuite/btcd/btcec) is integrated. Signatures
// and addresses produced with this curve will not be valid on the
// Ethereum network.
func S256() elliptic.Curve {
	return elliptic.P256()
}

// FromECDSA exports a private key into a binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	return priv.D.Bytes()
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	if len(d) == 0 {
		return nil, fmt.Errorf("private key bytes must not be empty")
	}

	curve := S256()
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = new(big.Int).SetBytes(d)

	// Validate that the key is within the curve order
	if priv.D.Cmp(curve.Params().N) >= 0 || priv.D.Sign() == 0 {
		return nil, fmt.Errorf("invalid private key: must be in range [1, N-1]")
	}

	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d)
	return priv, nil
}

// HexToECDSA parses a private key from a hex-encoded string.
// Accepts optional "0x" prefix.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	hexkey = strings.TrimPrefix(hexkey, "0x")
	if len(hexkey) == 0 {
		return nil, fmt.Errorf("empty hex key")
	}

	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	return ToECDSA(b)
}
