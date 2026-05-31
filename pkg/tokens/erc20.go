package tokens

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/khizar/bc-lib/internal/crypto"
	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/types"
)

// ERC20 represents an ERC20 token contract
type ERC20 struct {
	address types.Address
	chain   chains.Chain

	// Cached token info
	name     string
	symbol   string
	decimals uint8
	cached   bool
}

// Common ERC20 function signatures (first 4 bytes of keccak256 hash)
var (
	// Read functions
	fnName        = mustDecodeHex("06fdde03") // name()
	fnSymbol      = mustDecodeHex("95d89b41") // symbol()
	fnDecimals    = mustDecodeHex("313ce567") // decimals()
	fnTotalSupply = mustDecodeHex("18160ddd") // totalSupply()
	fnBalanceOf   = mustDecodeHex("70a08231") // balanceOf(address)
	fnAllowance   = mustDecodeHex("dd62ed3e") // allowance(address,address)

	// Write functions
	fnTransfer     = mustDecodeHex("a9059cbb") // transfer(address,uint256)
	fnApprove      = mustDecodeHex("095ea7b3") // approve(address,uint256)
	fnTransferFrom = mustDecodeHex("23b872dd") // transferFrom(address,address,uint256)
)

// NewERC20 creates a new ERC20 token instance
func NewERC20(address types.Address, chain chains.Chain) *ERC20 {
	return &ERC20{
		address: address,
		chain:   chain,
	}
}

// Address returns the token contract address
func (t *ERC20) Address() types.Address {
	return t.address
}

// Name returns the token name
func (t *ERC20) Name(ctx context.Context) (string, error) {
	if t.cached && t.name != "" {
		return t.name, nil
	}

	result, err := t.call(ctx, fnName)
	if err != nil {
		return "", fmt.Errorf("failed to get token name: %w", err)
	}

	t.name = decodeString(result)
	return t.name, nil
}

// Symbol returns the token symbol
func (t *ERC20) Symbol(ctx context.Context) (string, error) {
	if t.cached && t.symbol != "" {
		return t.symbol, nil
	}

	result, err := t.call(ctx, fnSymbol)
	if err != nil {
		return "", fmt.Errorf("failed to get token symbol: %w", err)
	}

	t.symbol = decodeString(result)
	return t.symbol, nil
}

// Decimals returns the token decimals
func (t *ERC20) Decimals(ctx context.Context) (uint8, error) {
	if t.cached {
		return t.decimals, nil
	}

	result, err := t.call(ctx, fnDecimals)
	if err != nil {
		return 0, fmt.Errorf("failed to get token decimals: %w", err)
	}

	if len(result) >= 32 {
		t.decimals = result[31]
	}
	t.cached = true
	return t.decimals, nil
}

// TotalSupply returns the total token supply
func (t *ERC20) TotalSupply(ctx context.Context) (*big.Int, error) {
	result, err := t.call(ctx, fnTotalSupply)
	if err != nil {
		return nil, fmt.Errorf("failed to get total supply: %w", err)
	}

	return new(big.Int).SetBytes(result), nil
}

// BalanceOf returns the token balance for an address
func (t *ERC20) BalanceOf(ctx context.Context, owner types.Address) (*big.Int, error) {
	data := append(fnBalanceOf, padAddress(owner)...)

	result, err := t.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return new(big.Int).SetBytes(result), nil
}

// Allowance returns the amount of tokens that spender is allowed to spend on behalf of owner
func (t *ERC20) Allowance(ctx context.Context, owner, spender types.Address) (*big.Int, error) {
	data := append(fnAllowance, padAddress(owner)...)
	data = append(data, padAddress(spender)...)

	result, err := t.call(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("failed to get allowance: %w", err)
	}

	return new(big.Int).SetBytes(result), nil
}

// TransferData returns the calldata for a transfer
func (t *ERC20) TransferData(to types.Address, amount *big.Int) []byte {
	data := append(fnTransfer, padAddress(to)...)
	data = append(data, padBigInt(amount)...)
	return data
}

// ApproveData returns the calldata for an approve
func (t *ERC20) ApproveData(spender types.Address, amount *big.Int) []byte {
	data := append(fnApprove, padAddress(spender)...)
	data = append(data, padBigInt(amount)...)
	return data
}

// TransferFromData returns the calldata for a transferFrom
func (t *ERC20) TransferFromData(from, to types.Address, amount *big.Int) []byte {
	data := append(fnTransferFrom, padAddress(from)...)
	data = append(data, padAddress(to)...)
	data = append(data, padBigInt(amount)...)
	return data
}

// FormatAmount formats an amount with the token's decimals
func (t *ERC20) FormatAmount(ctx context.Context, amount *big.Int) (string, error) {
	decimals, err := t.Decimals(ctx)
	if err != nil {
		return "", err
	}
	return FormatUnits(amount, decimals), nil
}

// ParseAmount parses a string amount to the token's base units
func (t *ERC20) ParseAmount(ctx context.Context, amount string) (*big.Int, error) {
	decimals, err := t.Decimals(ctx)
	if err != nil {
		return nil, err
	}
	return ParseUnits(amount, decimals)
}

// call performs a contract call
func (t *ERC20) call(ctx context.Context, data []byte) ([]byte, error) {
	msg := chains.CallMsg{
		To:   t.address,
		Data: data,
	}
	return t.chain.CallContract(ctx, msg, nil)
}

// Helper functions

func mustDecodeHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func padAddress(addr types.Address) []byte {
	padded := make([]byte, 32)
	copy(padded[12:], addr[:])
	return padded
}

func padBigInt(n *big.Int) []byte {
	if n == nil {
		return make([]byte, 32)
	}
	b := n.Bytes()
	if len(b) >= 32 {
		return b[:32]
	}
	padded := make([]byte, 32)
	copy(padded[32-len(b):], b)
	return padded
}

func decodeString(data []byte) string {
	if len(data) < 64 {
		// Try to decode as fixed bytes
		return strings.TrimRight(string(data), "\x00")
	}

	// Dynamic string: offset (32 bytes) + length (32 bytes) + data
	if len(data) >= 64 {
		length := new(big.Int).SetBytes(data[32:64]).Uint64()
		if len(data) >= 64+int(length) {
			return string(data[64 : 64+length])
		}
	}
	return ""
}

// FormatUnits formats a value with the given number of decimals
func FormatUnits(value *big.Int, decimals uint8) string {
	if value == nil {
		return "0"
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	quotient := new(big.Int).Div(value, divisor)
	remainder := new(big.Int).Mod(value, divisor)

	if remainder.Sign() == 0 {
		return quotient.String()
	}

	// Format with decimal places
	remainderStr := remainder.String()
	for len(remainderStr) < int(decimals) {
		remainderStr = "0" + remainderStr
	}
	remainderStr = strings.TrimRight(remainderStr, "0")

	return fmt.Sprintf("%s.%s", quotient.String(), remainderStr)
}

// ParseUnits parses a string value to the smallest unit
func ParseUnits(value string, decimals uint8) (*big.Int, error) {
	parts := strings.Split(value, ".")

	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid number format")
	}

	// Integer part
	intPart, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer part")
	}

	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	result := new(big.Int).Mul(intPart, multiplier)

	// Decimal part
	if len(parts) == 2 {
		decPart := parts[1]
		// Pad or trim to match decimals
		if len(decPart) > int(decimals) {
			decPart = decPart[:decimals]
		}
		for len(decPart) < int(decimals) {
			decPart += "0"
		}

		decValue, ok := new(big.Int).SetString(decPart, 10)
		if !ok {
			return nil, fmt.Errorf("invalid decimal part")
		}
		result.Add(result, decValue)
	}

	return result, nil
}

// CommonTokens contains addresses for well-known tokens
var CommonTokens = map[string]map[int64]types.Address{}

func init() {
	// USDT addresses
	CommonTokens["USDT"] = map[int64]types.Address{
		1:   types.MustAddressFromHex("0xdAC17F958D2ee523a2206206994597C13D831ec7"), // Ethereum
		137: types.MustAddressFromHex("0xc2132D05D31c914a87C6611C10748AEb04B58e8F"), // Polygon
	}

	// USDC addresses
	CommonTokens["USDC"] = map[int64]types.Address{
		1:   types.MustAddressFromHex("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // Ethereum
		137: types.MustAddressFromHex("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"), // Polygon
	}

	// WETH addresses
	CommonTokens["WETH"] = map[int64]types.Address{
		1:     types.MustAddressFromHex("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), // Ethereum
		137:   types.MustAddressFromHex("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619"), // Polygon
		42161: types.MustAddressFromHex("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1"), // Arbitrum
	}
}

// GetTokenAddress returns the address of a common token on a chain
func GetTokenAddress(symbol string, chainID int64) (types.Address, bool) {
	tokens, ok := CommonTokens[symbol]
	if !ok {
		return types.Address{}, false
	}
	addr, ok := tokens[chainID]
	return addr, ok
}

// TokenTransferEvent represents an ERC20 Transfer event
type TokenTransferEvent struct {
	From   types.Address
	To     types.Address
	Amount *big.Int
}

// TransferEventTopic is the keccak256 hash of Transfer(address,address,uint256)
var TransferEventTopic = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// ParseTransferEvent parses a Transfer event from a log
func ParseTransferEvent(log *types.Log) (*TokenTransferEvent, error) {
	if len(log.Topics) != 3 {
		return nil, fmt.Errorf("invalid number of topics for Transfer event")
	}

	if log.Topics[0] != TransferEventTopic {
		return nil, fmt.Errorf("not a Transfer event")
	}

	var from, to types.Address
	copy(from[:], log.Topics[1][12:])
	copy(to[:], log.Topics[2][12:])

	return &TokenTransferEvent{
		From:   from,
		To:     to,
		Amount: new(big.Int).SetBytes(log.Data),
	}, nil
}
