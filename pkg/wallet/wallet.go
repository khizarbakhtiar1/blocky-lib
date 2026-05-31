package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/khizar/bc-lib/internal/crypto"
	"github.com/khizar/bc-lib/internal/encoding"
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

// SignTransaction signs a transaction and returns a SignedTransaction whose
// RawTransaction field is ready to broadcast via SendRawTransaction.
//
// Legacy transactions are signed with EIP-155 replay protection; EIP-2930 and
// EIP-1559 transactions use the typed-transaction y-parity encoding. The input
// transaction is not mutated.
func (w *Wallet) SignTransaction(tx *types.Transaction) (*types.SignedTransaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("cannot sign nil transaction")
	}
	if tx.ChainID == nil || tx.ChainID.Sign() <= 0 {
		return nil, fmt.Errorf("transaction chain ID is required for signing")
	}

	// Work on a copy so the caller's transaction is left untouched.
	signed := *tx

	hash, err := encoding.SigningHash(&signed)
	if err != nil {
		return nil, fmt.Errorf("failed to compute signing hash: %w", err)
	}

	sig, err := crypto.Sign(hash[:], w.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	r := new(big.Int).SetBytes(sig[0:32])
	s := new(big.Int).SetBytes(sig[32:64])
	recid := int64(sig[64])

	switch signed.Type {
	case types.LegacyTxType:
		// EIP-155: v = recoveryID + 35 + chainID*2
		v := new(big.Int).Mul(signed.ChainID, big.NewInt(2))
		v.Add(v, big.NewInt(35+recid))
		signed.V = v
	default:
		// Typed transactions store the raw y-parity (0 or 1).
		signed.V = big.NewInt(recid)
	}
	signed.R = r
	signed.S = s

	raw, err := encoding.EncodeSigned(&signed)
	if err != nil {
		return nil, fmt.Errorf("failed to encode signed transaction: %w", err)
	}

	signed.From = w.address
	signed.Hash = encoding.TxHash(raw)

	return &types.SignedTransaction{
		Transaction:    &signed,
		RawTransaction: raw,
	}, nil
}

// String returns a string representation of the wallet (address only, never the key!)
func (w *Wallet) String() string {
	return w.address.String()
}
