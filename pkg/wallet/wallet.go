package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/khizar/bc-lib/internal/crypto"
	"github.com/khizar/bc-lib/pkg/types"
)

// Wallet represents a blockchain wallet with a private key
type Wallet struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	address    types.Address
}

// NewWallet creates a new wallet with a randomly generated private key
func NewWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	
	return newWalletFromKey(privateKey), nil
}

// FromPrivateKey creates a wallet from a private key
func FromPrivateKey(privateKey *ecdsa.PrivateKey) *Wallet {
	return newWalletFromKey(privateKey)
}

// FromPrivateKeyHex creates a wallet from a hex-encoded private key
func FromPrivateKeyHex(hexKey string) (*Wallet, error) {
	hexKey = strings.TrimPrefix(hexKey, "0x")
	
	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	
	privateKey, err := crypto.ToECDSA(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create private key: %w", err)
	}
	
	return newWalletFromKey(privateKey), nil
}

// newWalletFromKey creates a wallet from an existing private key
func newWalletFromKey(privateKey *ecdsa.PrivateKey) *Wallet {
	publicKey := &privateKey.PublicKey
	address := crypto.PubkeyToAddress(publicKey)
	
	return &Wallet{
		privateKey: privateKey,
		publicKey:  publicKey,
		address:    address,
	}
}

// Address returns the wallet's address
func (w *Wallet) Address() types.Address {
	return w.address
}

// PrivateKey returns the wallet's private key
func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

// PrivateKeyHex returns the private key as a hex string
func (w *Wallet) PrivateKeyHex() string {
	bytes := crypto.FromECDSA(w.privateKey)
	return hex.EncodeToString(bytes)
}

// PublicKey returns the wallet's public key
func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

// Sign signs a hash with the wallet's private key
func (w *Wallet) Sign(hash []byte) ([]byte, error) {
	return crypto.Sign(hash, w.privateKey)
}

// SignTransaction signs a transaction (to be implemented with proper RLP encoding)
func (w *Wallet) SignTransaction(tx *types.Transaction) (*types.SignedTransaction, error) {
	// This is a placeholder - full implementation requires RLP encoding
	// and proper EIP-155/EIP-1559 signature format
	return nil, fmt.Errorf("transaction signing not fully implemented yet")
}

// String returns a string representation of the wallet (address only, never the key!)
func (w *Wallet) String() string {
	return w.address.String()
}

