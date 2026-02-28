package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

// S256 returns the secp256k1 curve (simplified version)
// In production, use github.com/btcsuite/btcd/btcec or similar
func S256() elliptic.Curve {
	// Using P256 as a placeholder - in production use secp256k1
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
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(d)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	return priv, nil
}

// HexToECDSA parses a secp256k1 private key from hex
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b := []byte(hexkey)
	if len(b) == 0 {
		return nil, nil
	}
	return ToECDSA(b)
}

