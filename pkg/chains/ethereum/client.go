package ethereum

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/khizar/bc-lib/internal/rpc"
	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/types"
)

// Client implements the Chain interface for Ethereum and EVM-compatible chains
type Client struct {
	provider *rpc.Provider
	chainID  *big.Int
	name     string
}

// NewClient creates a new Ethereum client
func NewClient(config chains.ChainConfig) (*Client, error) {
	if len(config.RPCURLs) == 0 {
		return nil, fmt.Errorf("at least one RPC URL is required")
	}
	
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	
	provider := rpc.NewProvider(config.RPCURLs, timeout)
	
	return &Client{
		provider: provider,
		chainID:  config.ChainID,
		name:     config.Name,
	}, nil
}

// Connect establishes connection and verifies chain ID
func (c *Client) Connect(ctx context.Context) error {
	// Verify chain ID
	chainID, err := c.NetworkID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}
	
	if c.chainID != nil && chainID.Cmp(c.chainID) != 0 {
		return fmt.Errorf("chain ID mismatch: expected %s, got %s", c.chainID, chainID)
	}
	
	c.chainID = chainID
	
	// Set name if not provided
	if c.name == "" {
		c.name = chains.ChainName(c.chainID)
	}
	
	return nil
}

// Disconnect closes the connection
func (c *Client) Disconnect() error {
	return c.provider.Close()
}

// IsConnected returns true if connected
func (c *Client) IsConnected() bool {
	return c.provider != nil
}

// ChainID returns the chain ID
func (c *Client) ChainID() *big.Int {
	return c.chainID
}

// Name returns the chain name
func (c *Client) Name() string {
	return c.name
}

// GetBalance returns the balance of an address at the latest block
func (c *Client) GetBalance(ctx context.Context, address types.Address) (*big.Int, error) {
	return c.GetBalanceAt(ctx, address, nil)
}

// GetBalanceAt returns the balance of an address at a specific block
func (c *Client) GetBalanceAt(ctx context.Context, address types.Address, blockNumber *big.Int) (*big.Int, error) {
	blockParam := "latest"
	if blockNumber != nil {
		blockParam = fmt.Sprintf("0x%x", blockNumber)
	}
	
	result, err := c.provider.Call(ctx, "eth_getBalance", address.String(), blockParam)
	if err != nil {
		return nil, fmt.Errorf("eth_getBalance failed: %w", err)
	}
	
	var hexBalance string
	if err := json.Unmarshal(result, &hexBalance); err != nil {
		return nil, fmt.Errorf("failed to parse balance: %w", err)
	}
	
	balance := new(big.Int)
	if _, ok := balance.SetString(hexBalance[2:], 16); !ok {
		return nil, fmt.Errorf("invalid balance hex: %s", hexBalance)
	}
	
	return balance, nil
}

// GetBlockNumber returns the latest block number
func (c *Client) GetBlockNumber(ctx context.Context) (*big.Int, error) {
	result, err := c.provider.Call(ctx, "eth_blockNumber")
	if err != nil {
		return nil, fmt.Errorf("eth_blockNumber failed: %w", err)
	}
	
	var hexBlock string
	if err := json.Unmarshal(result, &hexBlock); err != nil {
		return nil, fmt.Errorf("failed to parse block number: %w", err)
	}
	
	blockNum := new(big.Int)
	if _, ok := blockNum.SetString(hexBlock[2:], 16); !ok {
		return nil, fmt.Errorf("invalid block number hex: %s", hexBlock)
	}
	
	return blockNum, nil
}

// GetTransactionCount returns the nonce for an address
func (c *Client) GetTransactionCount(ctx context.Context, address types.Address) (uint64, error) {
	return c.GetTransactionCountAt(ctx, address, nil)
}

// GetTransactionCountAt returns the nonce for an address at a specific block
func (c *Client) GetTransactionCountAt(ctx context.Context, address types.Address, blockNumber *big.Int) (uint64, error) {
	blockParam := "latest"
	if blockNumber != nil {
		blockParam = fmt.Sprintf("0x%x", blockNumber)
	}
	
	result, err := c.provider.Call(ctx, "eth_getTransactionCount", address.String(), blockParam)
	if err != nil {
		return 0, fmt.Errorf("eth_getTransactionCount failed: %w", err)
	}
	
	var hexNonce string
	if err := json.Unmarshal(result, &hexNonce); err != nil {
		return 0, fmt.Errorf("failed to parse nonce: %w", err)
	}
	
	nonce := new(big.Int)
	if _, ok := nonce.SetString(hexNonce[2:], 16); !ok {
		return 0, fmt.Errorf("invalid nonce hex: %s", hexNonce)
	}
	
	return nonce.Uint64(), nil
}

// SendRawTransaction sends a signed transaction
func (c *Client) SendRawTransaction(ctx context.Context, signedTx []byte) (types.Hash, error) {
	hexTx := fmt.Sprintf("0x%x", signedTx)
	
	result, err := c.provider.Call(ctx, "eth_sendRawTransaction", hexTx)
	if err != nil {
		return types.Hash{}, fmt.Errorf("eth_sendRawTransaction failed: %w", err)
	}
	
	var txHash string
	if err := json.Unmarshal(result, &txHash); err != nil {
		return types.Hash{}, fmt.Errorf("failed to parse tx hash: %w", err)
	}
	
	return types.HashFromHex(txHash)
}

// SendTransaction signs and sends a transaction (requires wallet integration)
func (c *Client) SendTransaction(ctx context.Context, tx *types.Transaction) (types.Hash, error) {
	return types.Hash{}, fmt.Errorf("SendTransaction requires wallet integration - use SendRawTransaction instead")
}

// EstimateGas estimates the gas needed for a transaction
func (c *Client) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
	callMsg := map[string]interface{}{
		"to":   tx.To.String(),
		"data": fmt.Sprintf("0x%x", tx.Data),
	}
	
	if tx.Value != nil && tx.Value.Sign() > 0 {
		callMsg["value"] = fmt.Sprintf("0x%x", tx.Value)
	}
	
	result, err := c.provider.Call(ctx, "eth_estimateGas", callMsg)
	if err != nil {
		return 0, fmt.Errorf("eth_estimateGas failed: %w", err)
	}
	
	var hexGas string
	if err := json.Unmarshal(result, &hexGas); err != nil {
		return 0, fmt.Errorf("failed to parse gas: %w", err)
	}
	
	gas := new(big.Int)
	if _, ok := gas.SetString(hexGas[2:], 16); !ok {
		return 0, fmt.Errorf("invalid gas hex: %s", hexGas)
	}
	
	return gas.Uint64(), nil
}

// SuggestGasPrice returns the suggested gas price
func (c *Client) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	result, err := c.provider.Call(ctx, "eth_gasPrice")
	if err != nil {
		return nil, fmt.Errorf("eth_gasPrice failed: %w", err)
	}
	
	var hexPrice string
	if err := json.Unmarshal(result, &hexPrice); err != nil {
		return nil, fmt.Errorf("failed to parse gas price: %w", err)
	}
	
	price := new(big.Int)
	if _, ok := price.SetString(hexPrice[2:], 16); !ok {
		return nil, fmt.Errorf("invalid gas price hex: %s", hexPrice)
	}
	
	return price, nil
}

// SuggestGasTipCap returns the suggested priority fee (EIP-1559)
func (c *Client) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	result, err := c.provider.Call(ctx, "eth_maxPriorityFeePerGas")
	if err != nil {
		return nil, fmt.Errorf("eth_maxPriorityFeePerGas failed: %w", err)
	}
	
	var hexTip string
	if err := json.Unmarshal(result, &hexTip); err != nil {
		return nil, fmt.Errorf("failed to parse tip: %w", err)
	}
	
	tip := new(big.Int)
	if _, ok := tip.SetString(hexTip[2:], 16); !ok {
		return nil, fmt.Errorf("invalid tip hex: %s", hexTip)
	}
	
	return tip, nil
}

// NetworkID returns the network ID
func (c *Client) NetworkID(ctx context.Context) (*big.Int, error) {
	result, err := c.provider.Call(ctx, "eth_chainId")
	if err != nil {
		return nil, fmt.Errorf("eth_chainId failed: %w", err)
	}
	
	var hexID string
	if err := json.Unmarshal(result, &hexID); err != nil {
		return nil, fmt.Errorf("failed to parse chain ID: %w", err)
	}
	
	id := new(big.Int)
	if _, ok := id.SetString(hexID[2:], 16); !ok {
		return nil, fmt.Errorf("invalid chain ID hex: %s", hexID)
	}
	
	return id, nil
}

// Placeholder implementations for remaining interface methods
// These will be implemented in future iterations

func (c *Client) GetTransaction(ctx context.Context, hash types.Hash) (*types.Transaction, error) {
	result, err := c.provider.Call(ctx, "eth_getTransactionByHash", hash.String())
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionByHash failed: %w", err)
	}
	
	var txData struct {
		Hash             string `json:"hash"`
		Nonce            string `json:"nonce"`
		From             string `json:"from"`
		To               string `json:"to"`
		Value            string `json:"value"`
		Gas              string `json:"gas"`
		GasPrice         string `json:"gasPrice"`
		MaxFeePerGas     string `json:"maxFeePerGas"`
		MaxPriorityFee   string `json:"maxPriorityFeePerGas"`
		Input            string `json:"input"`
		Type             string `json:"type"`
		ChainId          string `json:"chainId"`
		V                string `json:"v"`
		R                string `json:"r"`
		S                string `json:"s"`
	}
	
	if err := json.Unmarshal(result, &txData); err != nil {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}
	
	// Check if transaction was found
	if txData.Hash == "" {
		return nil, fmt.Errorf("transaction not found")
	}
	
	tx := &types.Transaction{}
	tx.Hash, _ = types.HashFromHex(txData.Hash)
	tx.Nonce = hexToUint64(txData.Nonce)
	tx.From, _ = types.AddressFromHex(txData.From)
	if txData.To != "" {
		to, _ := types.AddressFromHex(txData.To)
		tx.To = &to
	}
	tx.Value = hexToBigInt(txData.Value)
	tx.GasLimit = hexToUint64(txData.Gas)
	tx.GasPrice = hexToBigInt(txData.GasPrice)
	tx.MaxFeePerGas = hexToBigInt(txData.MaxFeePerGas)
	tx.MaxPriorityFeePerGas = hexToBigInt(txData.MaxPriorityFee)
	tx.Data = hexToBytes(txData.Input)
	tx.ChainID = hexToBigInt(txData.ChainId)
	tx.V = hexToBigInt(txData.V)
	tx.R = hexToBigInt(txData.R)
	tx.S = hexToBigInt(txData.S)
	
	// Set transaction type
	txType := hexToUint64(txData.Type)
	tx.Type = types.TransactionType(txType)
	
	return tx, nil
}

func (c *Client) GetTransactionReceipt(ctx context.Context, hash types.Hash) (*types.Receipt, error) {
	result, err := c.provider.Call(ctx, "eth_getTransactionReceipt", hash.String())
	if err != nil {
		return nil, fmt.Errorf("eth_getTransactionReceipt failed: %w", err)
	}
	
	var receiptData struct {
		TransactionHash   string `json:"transactionHash"`
		TransactionIndex  string `json:"transactionIndex"`
		BlockHash         string `json:"blockHash"`
		BlockNumber       string `json:"blockNumber"`
		From              string `json:"from"`
		To                string `json:"to"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
		GasUsed           string `json:"gasUsed"`
		ContractAddress   string `json:"contractAddress"`
		Status            string `json:"status"`
		EffectiveGasPrice string `json:"effectiveGasPrice"`
		Type              string `json:"type"`
		Logs              []struct {
			Address          string   `json:"address"`
			Topics           []string `json:"topics"`
			Data             string   `json:"data"`
			BlockNumber      string   `json:"blockNumber"`
			TransactionHash  string   `json:"transactionHash"`
			TransactionIndex string   `json:"transactionIndex"`
			BlockHash        string   `json:"blockHash"`
			LogIndex         string   `json:"logIndex"`
			Removed          bool     `json:"removed"`
		} `json:"logs"`
	}
	
	if err := json.Unmarshal(result, &receiptData); err != nil {
		return nil, fmt.Errorf("failed to parse receipt: %w", err)
	}
	
	if receiptData.TransactionHash == "" {
		return nil, fmt.Errorf("receipt not found")
	}
	
	receipt := &types.Receipt{
		TransactionHash:   types.MustHashFromHex(receiptData.TransactionHash),
		TransactionIndex:  hexToUint64(receiptData.TransactionIndex),
		BlockHash:         types.MustHashFromHex(receiptData.BlockHash),
		BlockNumber:       hexToBigInt(receiptData.BlockNumber),
		CumulativeGasUsed: hexToUint64(receiptData.CumulativeGasUsed),
		GasUsed:           hexToUint64(receiptData.GasUsed),
		Status:            hexToUint64(receiptData.Status),
		EffectiveGasPrice: hexToBigInt(receiptData.EffectiveGasPrice),
		Type:              uint8(hexToUint64(receiptData.Type)),
	}
	
	receipt.From, _ = types.AddressFromHex(receiptData.From)
	if receiptData.To != "" {
		to, _ := types.AddressFromHex(receiptData.To)
		receipt.To = &to
	}
	if receiptData.ContractAddress != "" {
		ca, _ := types.AddressFromHex(receiptData.ContractAddress)
		receipt.ContractAddress = &ca
	}
	
	// Parse logs
	for _, logData := range receiptData.Logs {
		log := &types.Log{
			BlockNumber:      hexToBigInt(logData.BlockNumber),
			BlockHash:        types.MustHashFromHex(logData.BlockHash),
			TransactionHash:  types.MustHashFromHex(logData.TransactionHash),
			TransactionIndex: uint(hexToUint64(logData.TransactionIndex)),
			LogIndex:         uint(hexToUint64(logData.LogIndex)),
			Removed:          logData.Removed,
			Data:             hexToBytes(logData.Data),
		}
		log.Address, _ = types.AddressFromHex(logData.Address)
		
		for _, topic := range logData.Topics {
			log.Topics = append(log.Topics, types.MustHashFromHex(topic))
		}
		
		receipt.Logs = append(receipt.Logs, log)
	}
	
	return receipt, nil
}

func (c *Client) GetBlockByNumber(ctx context.Context, blockNumber *big.Int, fullTx bool) (*types.Block, error) {
	blockParam := "latest"
	if blockNumber != nil {
		blockParam = fmt.Sprintf("0x%x", blockNumber)
	}
	
	result, err := c.provider.Call(ctx, "eth_getBlockByNumber", blockParam, fullTx)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByNumber failed: %w", err)
	}
	
	return c.parseBlock(result, fullTx)
}

func (c *Client) GetBlockByHash(ctx context.Context, hash types.Hash, fullTx bool) (*types.Block, error) {
	result, err := c.provider.Call(ctx, "eth_getBlockByHash", hash.String(), fullTx)
	if err != nil {
		return nil, fmt.Errorf("eth_getBlockByHash failed: %w", err)
	}
	
	return c.parseBlock(result, fullTx)
}

func (c *Client) parseBlock(data json.RawMessage, fullTx bool) (*types.Block, error) {
	var blockData struct {
		Number           string   `json:"number"`
		Hash             string   `json:"hash"`
		ParentHash       string   `json:"parentHash"`
		Nonce            string   `json:"nonce"`
		Sha3Uncles       string   `json:"sha3Uncles"`
		TransactionsRoot string   `json:"transactionsRoot"`
		StateRoot        string   `json:"stateRoot"`
		ReceiptsRoot     string   `json:"receiptsRoot"`
		Miner            string   `json:"miner"`
		Difficulty       string   `json:"difficulty"`
		TotalDifficulty  string   `json:"totalDifficulty"`
		ExtraData        string   `json:"extraData"`
		Size             string   `json:"size"`
		GasLimit         string   `json:"gasLimit"`
		GasUsed          string   `json:"gasUsed"`
		Timestamp        string   `json:"timestamp"`
		BaseFeePerGas    string   `json:"baseFeePerGas"`
		Transactions     []json.RawMessage `json:"transactions"`
		Uncles           []string `json:"uncles"`
	}
	
	if err := json.Unmarshal(data, &blockData); err != nil {
		return nil, fmt.Errorf("failed to parse block: %w", err)
	}
	
	if blockData.Hash == "" {
		return nil, fmt.Errorf("block not found")
	}
	
	block := &types.Block{
		Number:           hexToBigInt(blockData.Number),
		Hash:             types.MustHashFromHex(blockData.Hash),
		ParentHash:       types.MustHashFromHex(blockData.ParentHash),
		Nonce:            hexToUint64(blockData.Nonce),
		Sha3Uncles:       types.MustHashFromHex(blockData.Sha3Uncles),
		TransactionsRoot: types.MustHashFromHex(blockData.TransactionsRoot),
		StateRoot:        types.MustHashFromHex(blockData.StateRoot),
		ReceiptsRoot:     types.MustHashFromHex(blockData.ReceiptsRoot),
		Difficulty:       hexToBigInt(blockData.Difficulty),
		TotalDifficulty:  hexToBigInt(blockData.TotalDifficulty),
		ExtraData:        hexToBytes(blockData.ExtraData),
		Size:             hexToUint64(blockData.Size),
		GasLimit:         hexToUint64(blockData.GasLimit),
		GasUsed:          hexToUint64(blockData.GasUsed),
		Timestamp:        hexToUint64(blockData.Timestamp),
		BaseFeePerGas:    hexToBigInt(blockData.BaseFeePerGas),
	}
	
	block.Miner, _ = types.AddressFromHex(blockData.Miner)
	
	// Parse uncles
	for _, uncle := range blockData.Uncles {
		block.Uncles = append(block.Uncles, types.MustHashFromHex(uncle))
	}
	
	// Parse transactions
	for _, txRaw := range blockData.Transactions {
		if fullTx {
			// Full transaction object - would need to parse
			// For now, just extract hash
			var txObj struct {
				Hash string `json:"hash"`
			}
			if err := json.Unmarshal(txRaw, &txObj); err == nil {
				block.Transactions = append(block.Transactions, types.MustHashFromHex(txObj.Hash))
			}
		} else {
			// Just hash string
			var txHash string
			if err := json.Unmarshal(txRaw, &txHash); err == nil {
				block.Transactions = append(block.Transactions, types.MustHashFromHex(txHash))
			}
		}
	}
	
	return block, nil
}

func (c *Client) CallContract(ctx context.Context, msg chains.CallMsg, blockNumber *big.Int) ([]byte, error) {
	blockParam := "latest"
	if blockNumber != nil {
		blockParam = fmt.Sprintf("0x%x", blockNumber)
	}
	
	callMsg := map[string]interface{}{
		"to": msg.To.String(),
	}
	
	if !msg.From.IsZero() {
		callMsg["from"] = msg.From.String()
	}
	if msg.Gas > 0 {
		callMsg["gas"] = fmt.Sprintf("0x%x", msg.Gas)
	}
	if msg.GasPrice != nil {
		callMsg["gasPrice"] = fmt.Sprintf("0x%x", msg.GasPrice)
	}
	if msg.Value != nil && msg.Value.Sign() > 0 {
		callMsg["value"] = fmt.Sprintf("0x%x", msg.Value)
	}
	if len(msg.Data) > 0 {
		callMsg["data"] = fmt.Sprintf("0x%x", msg.Data)
	}
	
	result, err := c.provider.Call(ctx, "eth_call", callMsg, blockParam)
	if err != nil {
		return nil, fmt.Errorf("eth_call failed: %w", err)
	}
	
	var hexResult string
	if err := json.Unmarshal(result, &hexResult); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}
	
	return hexToBytes(hexResult), nil
}

func (c *Client) FilterLogs(ctx context.Context, query chains.LogQuery) ([]*types.Log, error) {
	filter := make(map[string]interface{})
	
	if query.FromBlock != nil {
		filter["fromBlock"] = fmt.Sprintf("0x%x", query.FromBlock)
	}
	if query.ToBlock != nil {
		filter["toBlock"] = fmt.Sprintf("0x%x", query.ToBlock)
	}
	if len(query.Addresses) > 0 {
		addrs := make([]string, len(query.Addresses))
		for i, a := range query.Addresses {
			addrs[i] = a.String()
		}
		filter["address"] = addrs
	}
	if len(query.Topics) > 0 {
		topics := make([]interface{}, len(query.Topics))
		for i, t := range query.Topics {
			if len(t) == 0 {
				topics[i] = nil
			} else if len(t) == 1 {
				topics[i] = t[0].String()
			} else {
				hashes := make([]string, len(t))
				for j, h := range t {
					hashes[j] = h.String()
				}
				topics[i] = hashes
			}
		}
		filter["topics"] = topics
	}
	
	result, err := c.provider.Call(ctx, "eth_getLogs", filter)
	if err != nil {
		return nil, fmt.Errorf("eth_getLogs failed: %w", err)
	}
	
	var logsData []struct {
		Address          string   `json:"address"`
		Topics           []string `json:"topics"`
		Data             string   `json:"data"`
		BlockNumber      string   `json:"blockNumber"`
		TransactionHash  string   `json:"transactionHash"`
		TransactionIndex string   `json:"transactionIndex"`
		BlockHash        string   `json:"blockHash"`
		LogIndex         string   `json:"logIndex"`
		Removed          bool     `json:"removed"`
	}
	
	if err := json.Unmarshal(result, &logsData); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}
	
	logs := make([]*types.Log, len(logsData))
	for i, logData := range logsData {
		log := &types.Log{
			BlockNumber:      hexToBigInt(logData.BlockNumber),
			BlockHash:        types.MustHashFromHex(logData.BlockHash),
			TransactionHash:  types.MustHashFromHex(logData.TransactionHash),
			TransactionIndex: uint(hexToUint64(logData.TransactionIndex)),
			LogIndex:         uint(hexToUint64(logData.LogIndex)),
			Removed:          logData.Removed,
			Data:             hexToBytes(logData.Data),
		}
		log.Address, _ = types.AddressFromHex(logData.Address)
		
		for _, topic := range logData.Topics {
			log.Topics = append(log.Topics, types.MustHashFromHex(topic))
		}
		
		logs[i] = log
	}
	
	return logs, nil
}

func (c *Client) SubscribeToLogs(ctx context.Context, query chains.LogQuery) (<-chan *types.Log, error) {
	return nil, fmt.Errorf("WebSocket subscription not implemented yet - use FilterLogs for polling")
}

func (c *Client) SyncProgress(ctx context.Context) (*chains.SyncProgress, error) {
	result, err := c.provider.Call(ctx, "eth_syncing")
	if err != nil {
		return nil, fmt.Errorf("eth_syncing failed: %w", err)
	}
	
	// If not syncing, returns false
	var syncing bool
	if err := json.Unmarshal(result, &syncing); err == nil && !syncing {
		return &chains.SyncProgress{Syncing: false}, nil
	}
	
	// Otherwise, returns sync progress object
	var syncData struct {
		StartingBlock string `json:"startingBlock"`
		CurrentBlock  string `json:"currentBlock"`
		HighestBlock  string `json:"highestBlock"`
	}
	
	if err := json.Unmarshal(result, &syncData); err != nil {
		return nil, fmt.Errorf("failed to parse sync progress: %w", err)
	}
	
	return &chains.SyncProgress{
		StartingBlock: hexToUint64(syncData.StartingBlock),
		CurrentBlock:  hexToUint64(syncData.CurrentBlock),
		HighestBlock:  hexToUint64(syncData.HighestBlock),
		Syncing:       true,
	}, nil
}

// Helper functions for parsing hex values
func hexToUint64(s string) uint64 {
	if s == "" || s == "0x" {
		return 0
	}
	s = strings.TrimPrefix(s, "0x")
	n := new(big.Int)
	n.SetString(s, 16)
	return n.Uint64()
}

func hexToBigInt(s string) *big.Int {
	if s == "" || s == "0x" {
		return nil
	}
	s = strings.TrimPrefix(s, "0x")
	n := new(big.Int)
	n.SetString(s, 16)
	return n
}

func hexToBytes(s string) []byte {
	if s == "" || s == "0x" {
		return nil
	}
	s = strings.TrimPrefix(s, "0x")
	b, _ := hex.DecodeString(s)
	return b
}

