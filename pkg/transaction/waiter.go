package transaction

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/types"
)

// Waiter provides utilities for waiting on transaction confirmation
type Waiter struct {
	chain         chains.Chain
	pollInterval  time.Duration
	maxWait       time.Duration
	confirmations int
}

// WaiterOption configures the Waiter
type WaiterOption func(*Waiter)

// WithPollInterval sets the polling interval
func WithPollInterval(d time.Duration) WaiterOption {
	return func(w *Waiter) {
		w.pollInterval = d
	}
}

// WithMaxWait sets the maximum wait time
func WithMaxWait(d time.Duration) WaiterOption {
	return func(w *Waiter) {
		w.maxWait = d
	}
}

// WithConfirmations sets the number of confirmations to wait for
func WithConfirmations(n int) WaiterOption {
	return func(w *Waiter) {
		w.confirmations = n
	}
}

// NewWaiter creates a new transaction waiter
func NewWaiter(chain chains.Chain, opts ...WaiterOption) *Waiter {
	w := &Waiter{
		chain:         chain,
		pollInterval:  2 * time.Second,
		maxWait:       5 * time.Minute,
		confirmations: 1,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// WaitResult contains the result of waiting for a transaction
type WaitResult struct {
	Receipt       *types.Receipt
	Confirmations int
	Success       bool
}

// WaitForReceipt waits for a transaction receipt
func (w *Waiter) WaitForReceipt(ctx context.Context, txHash types.Hash) (*WaitResult, error) {
	// Create a timeout context
	ctx, cancel := context.WithTimeout(ctx, w.maxWait)
	defer cancel()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction %s: %w", txHash, ctx.Err())
		case <-ticker.C:
			receipt, err := w.chain.GetTransactionReceipt(ctx, txHash)
			if err != nil {
				// Transaction not mined yet, continue polling
				continue
			}

			// Check confirmations if required
			if w.confirmations > 1 {
				currentBlock, err := w.chain.GetBlockNumber(ctx)
				if err != nil {
					continue
				}

				confirmations := int(currentBlock.Int64() - receipt.BlockNumber.Int64() + 1)
				if confirmations < w.confirmations {
					continue
				}
			}

			return &WaitResult{
				Receipt:       receipt,
				Confirmations: w.confirmations,
				Success:       receipt.Status == 1,
			}, nil
		}
	}
}

// WaitForConfirmations waits for a specific number of confirmations
func (w *Waiter) WaitForConfirmations(ctx context.Context, txHash types.Hash, confirmations int) (*WaitResult, error) {
	// First wait for receipt
	result, err := w.WaitForReceipt(ctx, txHash)
	if err != nil {
		return nil, err
	}

	if confirmations <= 1 {
		return result, nil
	}

	// Wait for additional confirmations
	ctx, cancel := context.WithTimeout(ctx, w.maxWait)
	defer cancel()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for confirmations: %w", ctx.Err())
		case <-ticker.C:
			currentBlock, err := w.chain.GetBlockNumber(ctx)
			if err != nil {
				continue
			}

			currentConfirmations := int(currentBlock.Int64() - result.Receipt.BlockNumber.Int64() + 1)
			if currentConfirmations >= confirmations {
				result.Confirmations = currentConfirmations
				return result, nil
			}
		}
	}
}

// WaitForBlock waits for a specific block number
func WaitForBlock(ctx context.Context, chain chains.Chain, targetBlock int64, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			currentBlock, err := chain.GetBlockNumber(ctx)
			if err != nil {
				continue
			}

			if currentBlock.Int64() >= targetBlock {
				return nil
			}
		}
	}
}

// TransactionStatus represents the status of a transaction
type TransactionStatus int

const (
	StatusUnknown TransactionStatus = iota
	StatusPending
	StatusMined
	StatusConfirmed
	StatusFailed
)

func (s TransactionStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusMined:
		return "mined"
	case StatusConfirmed:
		return "confirmed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// GetTransactionStatus checks the current status of a transaction
func GetTransactionStatus(ctx context.Context, chain chains.Chain, txHash types.Hash, requiredConfirmations int) (TransactionStatus, int, error) {
	// Try to get the transaction
	tx, err := chain.GetTransaction(ctx, txHash)
	if err != nil {
		return StatusUnknown, 0, nil // Transaction not found
	}

	// Transaction exists, check for receipt
	receipt, err := chain.GetTransactionReceipt(ctx, txHash)
	if err != nil {
		return StatusPending, 0, nil // Transaction pending
	}

	// Check if transaction failed
	if receipt.Status == 0 {
		return StatusFailed, 0, nil
	}

	// Get current block for confirmation count
	currentBlock, err := chain.GetBlockNumber(ctx)
	if err != nil {
		return StatusMined, 1, nil
	}

	confirmations := int(currentBlock.Int64() - receipt.BlockNumber.Int64() + 1)

	if confirmations >= requiredConfirmations {
		return StatusConfirmed, confirmations, nil
	}

	// Silence the unused variable warning
	_ = tx

	return StatusMined, confirmations, nil
}

// SpeedUpTransaction creates a replacement transaction with higher gas price
func SpeedUpTransaction(original *types.Transaction, gasPriceIncrease int) *types.Transaction {
	replacement := *original // Copy

	if replacement.Type == types.DynamicFeeTxType {
		// EIP-1559 transaction
		if replacement.MaxFeePerGas != nil {
			increase := int64(gasPriceIncrease)
			newMaxFee := new(big.Int).Mul(replacement.MaxFeePerGas, big.NewInt(100+increase))
			newMaxFee.Div(newMaxFee, big.NewInt(100))
			replacement.MaxFeePerGas = newMaxFee
		}
		if replacement.MaxPriorityFeePerGas != nil {
			increase := int64(gasPriceIncrease)
			newPriorityFee := new(big.Int).Mul(replacement.MaxPriorityFeePerGas, big.NewInt(100+increase))
			newPriorityFee.Div(newPriorityFee, big.NewInt(100))
			replacement.MaxPriorityFeePerGas = newPriorityFee
		}
	} else {
		// Legacy transaction
		if replacement.GasPrice != nil {
			increase := int64(gasPriceIncrease)
			newGasPrice := new(big.Int).Mul(replacement.GasPrice, big.NewInt(100+increase))
			newGasPrice.Div(newGasPrice, big.NewInt(100))
			replacement.GasPrice = newGasPrice
		}
	}

	return &replacement
}

// CancelTransaction creates a cancellation transaction (0 value to self)
func CancelTransaction(from types.Address, nonce uint64, chainID *big.Int, maxFeePerGas, maxPriorityFeePerGas *big.Int) *types.Transaction {
	return &types.Transaction{
		Type:                 types.DynamicFeeTxType,
		ChainID:              chainID,
		Nonce:                nonce,
		To:                   &from, // Send to self
		Value:                big.NewInt(0),
		GasLimit:             21000,
		MaxFeePerGas:         maxFeePerGas,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		Data:                 []byte{},
	}
}
