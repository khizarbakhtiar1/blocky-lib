package errors

import (
	"errors"
	"fmt"
)

// Standard error types for the blockchain library
var (
	// Connection errors
	ErrNotConnected     = errors.New("not connected to blockchain")
	ErrConnectionFailed = errors.New("failed to connect to blockchain")
	ErrAllRPCsFailed    = errors.New("all RPC endpoints failed")
	ErrTimeout          = errors.New("request timed out")

	// Transaction errors
	ErrInsufficientFunds      = errors.New("insufficient funds for transaction")
	ErrNonceTooLow            = errors.New("nonce too low")
	ErrNonceTooHigh           = errors.New("nonce too high")
	ErrGasTooLow              = errors.New("gas limit too low")
	ErrGasPriceTooLow         = errors.New("gas price too low")
	ErrTransactionReverted    = errors.New("transaction reverted")
	ErrTransactionUnderpriced = errors.New("transaction underpriced")
	ErrReplacementUnderpriced = errors.New("replacement transaction underpriced")

	// Validation errors
	ErrInvalidAddress     = errors.New("invalid address")
	ErrInvalidHash        = errors.New("invalid hash")
	ErrInvalidSignature   = errors.New("invalid signature")
	ErrInvalidPrivateKey  = errors.New("invalid private key")
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrInvalidChainID     = errors.New("invalid chain ID")

	// State errors
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrReceiptNotFound     = errors.New("transaction receipt not found")
	ErrAccountNotFound     = errors.New("account not found")

	// Contract errors
	ErrContractExecution = errors.New("contract execution failed")
	ErrContractNotFound  = errors.New("contract not found")
	ErrABIEncodingFailed = errors.New("ABI encoding failed")
	ErrABIDecodingFailed = errors.New("ABI decoding failed")
)

// RPCError represents an error from the RPC provider
type RPCError struct {
	Code    int
	Message string
	Data    interface{}
}

func (e *RPCError) Error() string {
	if e.Data != nil {
		return fmt.Sprintf("RPC error %d: %s (data: %v)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("RPC error %d: %s", e.Code, e.Message)
}

// IsRPCError checks if an error is an RPCError
func IsRPCError(err error) bool {
	var rpcErr *RPCError
	return errors.As(err, &rpcErr)
}

// TransactionError wraps transaction-related errors with additional context
type TransactionError struct {
	TxHash  string
	Message string
	Cause   error
}

func (e *TransactionError) Error() string {
	if e.TxHash != "" {
		return fmt.Sprintf("transaction %s: %s", e.TxHash, e.Message)
	}
	return e.Message
}

func (e *TransactionError) Unwrap() error {
	return e.Cause
}

// NewTransactionError creates a new TransactionError
func NewTransactionError(hash, message string, cause error) *TransactionError {
	return &TransactionError{
		TxHash:  hash,
		Message: message,
		Cause:   cause,
	}
}

// ChainError represents chain-specific errors
type ChainError struct {
	ChainID   int64
	ChainName string
	Message   string
	Cause     error
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("chain %s (%d): %s", e.ChainName, e.ChainID, e.Message)
}

func (e *ChainError) Unwrap() error {
	return e.Cause
}

// NewChainError creates a new ChainError
func NewChainError(chainID int64, chainName, message string, cause error) *ChainError {
	return &ChainError{
		ChainID:   chainID,
		ChainName: chainName,
		Message:   message,
		Cause:     cause,
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Is checks if an error matches a target
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
