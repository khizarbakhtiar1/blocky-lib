package types

import (
	"math/big"
)

// TransactionType represents different transaction types
type TransactionType uint8

const (
	// LegacyTxType is the original transaction type
	LegacyTxType TransactionType = 0x00
	// AccessListTxType is EIP-2930
	AccessListTxType TransactionType = 0x01
	// DynamicFeeTxType is EIP-1559
	DynamicFeeTxType TransactionType = 0x02
)

// Transaction represents a blockchain transaction
type Transaction struct {
	// Type of transaction (legacy, EIP-2930, EIP-1559)
	Type TransactionType

	// Common fields
	ChainID  *big.Int
	Nonce    uint64
	To       *Address // nil means contract creation
	Value    *big.Int
	Data     []byte
	GasLimit uint64

	// Legacy and EIP-2930 fields
	GasPrice *big.Int

	// EIP-1559 fields
	MaxFeePerGas         *big.Int
	MaxPriorityFeePerGas *big.Int

	// EIP-2930 access list
	AccessList AccessList

	// Signature fields
	V *big.Int
	R *big.Int
	S *big.Int

	// Computed fields (set after signing/sending)
	Hash Hash
	From Address
}

// AccessList is an EIP-2930 access list
type AccessList []AccessTuple

// AccessTuple represents a single access list entry
type AccessTuple struct {
	Address     Address
	StorageKeys []Hash
}

// NewTransaction creates a new transaction
func NewTransaction(nonce uint64, to Address, value *big.Int, gasLimit uint64, data []byte) *Transaction {
	return &Transaction{
		Type:     DynamicFeeTxType, // Default to EIP-1559
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Data:     data,
		GasLimit: gasLimit,
	}
}

// NewContractCreation creates a transaction for contract deployment
func NewContractCreation(nonce uint64, value *big.Int, gasLimit uint64, data []byte) *Transaction {
	return &Transaction{
		Type:     DynamicFeeTxType,
		Nonce:    nonce,
		To:       nil, // nil means contract creation
		Value:    value,
		Data:     data,
		GasLimit: gasLimit,
	}
}

// IsContractCreation returns true if this is a contract creation transaction
func (tx *Transaction) IsContractCreation() bool {
	return tx.To == nil
}

// Cost returns the total cost of the transaction (value + gas cost)
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int)
	if tx.Value != nil {
		total.Set(tx.Value)
	}

	gasCost := new(big.Int).SetUint64(tx.GasLimit)
	if tx.Type == DynamicFeeTxType && tx.MaxFeePerGas != nil {
		gasCost.Mul(gasCost, tx.MaxFeePerGas)
	} else if tx.GasPrice != nil {
		gasCost.Mul(gasCost, tx.GasPrice)
	}

	total.Add(total, gasCost)
	return total
}

// SignedTransaction represents a signed transaction ready to be broadcast
type SignedTransaction struct {
	*Transaction
	RawTransaction []byte // RLP-encoded transaction
}
