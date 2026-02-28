package client

import (
	"context"
	"fmt"
	"math/big"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/chains/ethereum"
	"github.com/khizar/bc-lib/pkg/types"
)

// Client is the main entry point for interacting with multiple blockchains
type Client struct {
	chains map[string]chains.Chain
	config Config
}

// Config holds the configuration for the multi-chain client
type Config struct {
	Chains map[string]chains.ChainConfig
}

// New creates a new multi-chain client
func New(config Config) (*Client, error) {
	if len(config.Chains) == 0 {
		return nil, fmt.Errorf("at least one chain configuration is required")
	}
	
	client := &Client{
		chains: make(map[string]chains.Chain),
		config: config,
	}
	
	// Initialize chains
	for name, chainConfig := range config.Chains {
		chain, err := createChain(chainConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain %s: %w", name, err)
		}
		client.chains[name] = chain
	}
	
	return client, nil
}

// createChain creates a chain client based on the chain ID
func createChain(config chains.ChainConfig) (chains.Chain, error) {
	// For now, all chains use the Ethereum client (EVM-compatible)
	// In the future, we'll add Solana, Cosmos, etc.
	return ethereum.NewClient(config)
}

// GetChain returns a chain client by name
func (c *Client) GetChain(name string) (chains.Chain, error) {
	chain, ok := c.chains[name]
	if !ok {
		return nil, fmt.Errorf("chain not found: %s", name)
	}
	return chain, nil
}

// Connect connects to a specific chain
func (c *Client) Connect(ctx context.Context, chainName string) error {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return err
	}
	return chain.Connect(ctx)
}

// ConnectAll connects to all configured chains
func (c *Client) ConnectAll(ctx context.Context) error {
	for name, chain := range c.chains {
		if err := chain.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", name, err)
		}
	}
	return nil
}

// GetBalance returns the balance of an address on a specific chain
func (c *Client) GetBalance(ctx context.Context, chainName string, address types.Address) (*big.Int, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return nil, err
	}
	return chain.GetBalance(ctx, address)
}

// GetBlockNumber returns the latest block number for a chain
func (c *Client) GetBlockNumber(ctx context.Context, chainName string) (*big.Int, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return nil, err
	}
	return chain.GetBlockNumber(ctx)
}

// GetNonce returns the transaction count (nonce) for an address
func (c *Client) GetNonce(ctx context.Context, chainName string, address types.Address) (uint64, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return 0, err
	}
	return chain.GetTransactionCount(ctx, address)
}

// SendRawTransaction sends a signed transaction to a specific chain
func (c *Client) SendRawTransaction(ctx context.Context, chainName string, signedTx []byte) (types.Hash, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return types.Hash{}, err
	}
	return chain.SendRawTransaction(ctx, signedTx)
}

// EstimateGas estimates the gas needed for a transaction
func (c *Client) EstimateGas(ctx context.Context, chainName string, tx *types.Transaction) (uint64, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return 0, err
	}
	return chain.EstimateGas(ctx, tx)
}

// SuggestGasPrice returns the suggested gas price for a chain
func (c *Client) SuggestGasPrice(ctx context.Context, chainName string) (*big.Int, error) {
	chain, err := c.GetChain(chainName)
	if err != nil {
		return nil, err
	}
	return chain.SuggestGasPrice(ctx)
}

// ChainNames returns all configured chain names
func (c *Client) ChainNames() []string {
	names := make([]string, 0, len(c.chains))
	for name := range c.chains {
		names = append(names, name)
	}
	return names
}

// Close closes all chain connections
func (c *Client) Close() error {
	for _, chain := range c.chains {
		if err := chain.Disconnect(); err != nil {
			return err
		}
	}
	return nil
}

