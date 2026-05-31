package chains

import (
	"context"
	"math/big"

	"github.com/khizar/bc-lib/pkg/types"
)

// Chain represents a blockchain interface that all chain implementations must satisfy
type Chain interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool
	ChainID() *big.Int
	Name() string

	// Balance queries
	GetBalance(ctx context.Context, address types.Address) (*big.Int, error)
	GetBalanceAt(ctx context.Context, address types.Address, blockNumber *big.Int) (*big.Int, error)

	// Transaction queries
	GetTransaction(ctx context.Context, hash types.Hash) (*types.Transaction, error)
	GetTransactionReceipt(ctx context.Context, hash types.Hash) (*types.Receipt, error)
	GetTransactionCount(ctx context.Context, address types.Address) (uint64, error)
	GetTransactionCountAt(ctx context.Context, address types.Address, blockNumber *big.Int) (uint64, error)

	// Block queries
	GetBlockNumber(ctx context.Context) (*big.Int, error)
	GetBlockByNumber(ctx context.Context, blockNumber *big.Int, fullTx bool) (*types.Block, error)
	GetBlockByHash(ctx context.Context, hash types.Hash, fullTx bool) (*types.Block, error)

	// Transaction submission
	SendRawTransaction(ctx context.Context, signedTx []byte) (types.Hash, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) (types.Hash, error)

	// Gas estimation
	EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)

	// Contract interaction
	CallContract(ctx context.Context, msg CallMsg, blockNumber *big.Int) ([]byte, error)

	// Event logs
	FilterLogs(ctx context.Context, query LogQuery) ([]*types.Log, error)
	SubscribeToLogs(ctx context.Context, query LogQuery) (<-chan *types.Log, error)

	// Network info
	NetworkID(ctx context.Context) (*big.Int, error)
	SyncProgress(ctx context.Context) (*SyncProgress, error)
}

// CallMsg represents a message call for contract interaction
type CallMsg struct {
	From     types.Address
	To       types.Address
	Gas      uint64
	GasPrice *big.Int
	Value    *big.Int
	Data     []byte
}

// LogQuery represents a filter for event logs
type LogQuery struct {
	FromBlock *big.Int
	ToBlock   *big.Int
	Addresses []types.Address
	Topics    [][]types.Hash
}

// SyncProgress represents the sync status of a node
type SyncProgress struct {
	StartingBlock uint64
	CurrentBlock  uint64
	HighestBlock  uint64
	Syncing       bool
}

// ChainConfig contains configuration for a blockchain connection
type ChainConfig struct {
	ChainID  *big.Int
	Name     string
	RPCURLs  []string
	WSURLs   []string
	Timeout  int // seconds
	MaxRetry int
}

// Common chain IDs
var (
	EthereumMainnet = big.NewInt(1)
	EthereumSepolia = big.NewInt(11155111)
	PolygonMainnet  = big.NewInt(137)
	PolygonMumbai   = big.NewInt(80001)
	ArbitrumOne     = big.NewInt(42161)
	ArbitrumSepolia = big.NewInt(421614)
	OptimismMainnet = big.NewInt(10)
	OptimismSepolia = big.NewInt(11155420)
	BaseMainnet     = big.NewInt(8453)
	BaseSepolia     = big.NewInt(84532)
)

// ChainName returns the name for common chain IDs
func ChainName(chainID *big.Int) string {
	switch chainID.Int64() {
	case 1:
		return "Ethereum Mainnet"
	case 11155111:
		return "Ethereum Sepolia"
	case 137:
		return "Polygon Mainnet"
	case 80001:
		return "Polygon Mumbai"
	case 42161:
		return "Arbitrum One"
	case 421614:
		return "Arbitrum Sepolia"
	case 10:
		return "Optimism Mainnet"
	case 11155420:
		return "Optimism Sepolia"
	case 8453:
		return "Base Mainnet"
	case 84532:
		return "Base Sepolia"
	default:
		return "Unknown Chain"
	}
}
