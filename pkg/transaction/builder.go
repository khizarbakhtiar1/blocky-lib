package transaction

import (
	"fmt"
	"math/big"

	"github.com/khizar/bc-lib/pkg/types"
)

// Builder helps construct transactions with a fluent API
type Builder struct {
	tx  *types.Transaction
	err error
}

// NewBuilder creates a new transaction builder
func NewBuilder() *Builder {
	return &Builder{
		tx: &types.Transaction{
			Type:     types.DynamicFeeTxType, // Default to EIP-1559
			Value:    big.NewInt(0),
			Data:     []byte{},
			GasLimit: 21000, // Default gas limit for simple transfer
		},
	}
}

// NewLegacyBuilder creates a builder for legacy transactions
func NewLegacyBuilder() *Builder {
	b := NewBuilder()
	b.tx.Type = types.LegacyTxType
	return b
}

// To sets the recipient address
func (b *Builder) To(address types.Address) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.To = &address
	return b
}

// Value sets the transaction value in wei
func (b *Builder) Value(value *big.Int) *Builder {
	if b.err != nil {
		return b
	}
	if value == nil {
		value = big.NewInt(0)
	}
	b.tx.Value = value
	return b
}

// ValueEther sets the transaction value in ether
func (b *Builder) ValueEther(ether float64) *Builder {
	if b.err != nil {
		return b
	}
	value, err := types.ParseEther(fmt.Sprintf("%.18f", ether))
	if err != nil {
		b.err = err
		return b
	}
	b.tx.Value = value
	return b
}

// Data sets the transaction data
func (b *Builder) Data(data []byte) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.Data = data
	return b
}

// Nonce sets the transaction nonce
func (b *Builder) Nonce(nonce uint64) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.Nonce = nonce
	return b
}

// GasLimit sets the gas limit
func (b *Builder) GasLimit(limit uint64) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.GasLimit = limit
	return b
}

// GasPrice sets the gas price (for legacy transactions)
func (b *Builder) GasPrice(price *big.Int) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.GasPrice = price
	b.tx.Type = types.LegacyTxType
	return b
}

// MaxFeePerGas sets the max fee per gas (EIP-1559)
func (b *Builder) MaxFeePerGas(fee *big.Int) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.MaxFeePerGas = fee
	b.tx.Type = types.DynamicFeeTxType
	return b
}

// MaxPriorityFeePerGas sets the max priority fee per gas (EIP-1559)
func (b *Builder) MaxPriorityFeePerGas(fee *big.Int) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.MaxPriorityFeePerGas = fee
	b.tx.Type = types.DynamicFeeTxType
	return b
}

// ChainID sets the chain ID
func (b *Builder) ChainID(chainID *big.Int) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.ChainID = chainID
	return b
}

// AccessList sets the access list (EIP-2930)
func (b *Builder) AccessList(accessList types.AccessList) *Builder {
	if b.err != nil {
		return b
	}
	b.tx.AccessList = accessList
	if b.tx.Type == types.LegacyTxType {
		b.tx.Type = types.AccessListTxType
	}
	return b
}

// Build builds the transaction
func (b *Builder) Build() (*types.Transaction, error) {
	if b.err != nil {
		return nil, b.err
	}
	
	// Validate transaction
	if err := b.validate(); err != nil {
		return nil, err
	}
	
	return b.tx, nil
}

// validate checks if the transaction is valid
func (b *Builder) validate() error {
	// Check chain ID
	if b.tx.ChainID == nil {
		return fmt.Errorf("chain ID is required")
	}
	
	// Check gas settings based on type
	if b.tx.Type == types.DynamicFeeTxType {
		if b.tx.MaxFeePerGas == nil {
			return fmt.Errorf("maxFeePerGas is required for EIP-1559 transactions")
		}
		if b.tx.MaxPriorityFeePerGas == nil {
			return fmt.Errorf("maxPriorityFeePerGas is required for EIP-1559 transactions")
		}
		if b.tx.MaxPriorityFeePerGas.Cmp(b.tx.MaxFeePerGas) > 0 {
			return fmt.Errorf("maxPriorityFeePerGas cannot exceed maxFeePerGas")
		}
	} else {
		if b.tx.GasPrice == nil {
			return fmt.Errorf("gasPrice is required for legacy transactions")
		}
	}
	
	// Check gas limit
	if b.tx.GasLimit == 0 {
		return fmt.Errorf("gas limit cannot be zero")
	}
	
	return nil
}

// MustBuild builds the transaction or panics
func (b *Builder) MustBuild() *types.Transaction {
	tx, err := b.Build()
	if err != nil {
		panic(err)
	}
	return tx
}

// Transfer creates a simple transfer transaction
func Transfer(to types.Address, value *big.Int, chainID *big.Int) *Builder {
	return NewBuilder().
		To(to).
		Value(value).
		ChainID(chainID).
		GasLimit(21000)
}

// ContractCall creates a contract call transaction
func ContractCall(to types.Address, data []byte, chainID *big.Int) *Builder {
	return NewBuilder().
		To(to).
		Data(data).
		ChainID(chainID)
}

// ContractDeploy creates a contract deployment transaction
func ContractDeploy(bytecode []byte, chainID *big.Int) *Builder {
	b := NewBuilder()
	b.tx.To = nil // nil means contract creation
	b.tx.Data = bytecode
	b.tx.ChainID = chainID
	return b
}

