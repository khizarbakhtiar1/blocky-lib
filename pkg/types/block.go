package types

import (
	"math/big"
	"time"
)

// Block represents a blockchain block
type Block struct {
	// Header fields
	Number           *big.Int
	Hash             Hash
	ParentHash       Hash
	Nonce            uint64
	Sha3Uncles       Hash
	TransactionsRoot Hash
	StateRoot        Hash
	ReceiptsRoot     Hash
	Miner            Address
	Difficulty       *big.Int
	TotalDifficulty  *big.Int
	ExtraData        []byte
	Size             uint64
	GasLimit         uint64
	GasUsed          uint64
	Timestamp        uint64

	// EIP-1559 fields
	BaseFeePerGas *big.Int

	// Transactions (can be hashes or full transactions)
	Transactions     []Hash
	TransactionsFull []*Transaction

	// Uncles
	Uncles []Hash
}

// Time returns the block timestamp as time.Time
func (b *Block) Time() time.Time {
	return time.Unix(int64(b.Timestamp), 0)
}

// HasTransactions returns true if the block contains transactions
func (b *Block) HasTransactions() bool {
	return len(b.Transactions) > 0 || len(b.TransactionsFull) > 0
}

// TransactionCount returns the number of transactions in the block
func (b *Block) TransactionCount() int {
	if len(b.TransactionsFull) > 0 {
		return len(b.TransactionsFull)
	}
	return len(b.Transactions)
}

// BlockHeader contains just the block header information
type BlockHeader struct {
	Number          *big.Int
	Hash            Hash
	ParentHash      Hash
	Timestamp       uint64
	Miner           Address
	GasLimit        uint64
	GasUsed         uint64
	BaseFeePerGas   *big.Int
	Difficulty      *big.Int
	TotalDifficulty *big.Int
}

// Log represents an event log from a smart contract
type Log struct {
	Address          Address
	Topics           []Hash
	Data             []byte
	BlockNumber      *big.Int
	BlockHash        Hash
	TransactionHash  Hash
	TransactionIndex uint
	LogIndex         uint
	Removed          bool // true if log was removed due to chain reorg
}

// Receipt represents a transaction receipt
type Receipt struct {
	TransactionHash   Hash
	TransactionIndex  uint64
	BlockHash         Hash
	BlockNumber       *big.Int
	From              Address
	To                *Address
	CumulativeGasUsed uint64
	GasUsed           uint64
	ContractAddress   *Address // set if contract creation
	Logs              []*Log
	Status            uint64 // 1 = success, 0 = failure
	EffectiveGasPrice *big.Int

	// EIP-1559
	Type uint8
}

// IsSuccess returns true if the transaction was successful
func (r *Receipt) IsSuccess() bool {
	return r.Status == 1
}

// IsContractCreation returns true if this receipt is for contract creation
func (r *Receipt) IsContractCreation() bool {
	return r.ContractAddress != nil
}
