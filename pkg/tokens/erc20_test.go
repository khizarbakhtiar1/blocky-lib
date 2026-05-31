package tokens

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatUnits(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		decimals uint8
		expected string
	}{
		{
			name:     "1 ETH",
			value:    big.NewInt(1000000000000000000),
			decimals: 18,
			expected: "1",
		},
		{
			name:     "1.5 ETH",
			value:    big.NewInt(1500000000000000000),
			decimals: 18,
			expected: "1.5",
		},
		{
			name:     "0.123456789 ETH",
			value:    big.NewInt(123456789000000000),
			decimals: 18,
			expected: "0.123456789",
		},
		{
			name:     "1 USDC (6 decimals)",
			value:    big.NewInt(1000000),
			decimals: 6,
			expected: "1",
		},
		{
			name:     "1.23 USDC",
			value:    big.NewInt(1230000),
			decimals: 6,
			expected: "1.23",
		},
		{
			name:     "zero",
			value:    big.NewInt(0),
			decimals: 18,
			expected: "0",
		},
		{
			name:     "nil",
			value:    nil,
			decimals: 18,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatUnits(tt.value, tt.decimals)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseUnits(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		decimals uint8
		expected *big.Int
		hasError bool
	}{
		{
			name:     "1 ETH",
			value:    "1",
			decimals: 18,
			expected: big.NewInt(1000000000000000000),
		},
		{
			name:     "1.5 ETH",
			value:    "1.5",
			decimals: 18,
			expected: big.NewInt(1500000000000000000),
		},
		{
			name:     "0.123 ETH",
			value:    "0.123",
			decimals: 18,
			expected: big.NewInt(123000000000000000),
		},
		{
			name:     "1 USDC",
			value:    "1",
			decimals: 6,
			expected: big.NewInt(1000000),
		},
		{
			name:     "0 value",
			value:    "0",
			decimals: 18,
			expected: big.NewInt(0),
		},
		{
			name:     "invalid format",
			value:    "1.2.3",
			decimals: 18,
			hasError: true,
		},
		{
			name:     "invalid integer part",
			value:    "abc",
			decimals: 18,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseUnits(tt.value, tt.decimals)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, 0, tt.expected.Cmp(result))
			}
		})
	}
}

func TestPadAddress(t *testing.T) {
	addr := mustDecodeHex("742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	var address [20]byte
	copy(address[:], addr)

	// Need to import the Address type properly for this test
	// For now, just test the helper functions
	result := padBigInt(big.NewInt(256))
	assert.Equal(t, 32, len(result))
	assert.Equal(t, byte(1), result[30])
	assert.Equal(t, byte(0), result[31])
}

func TestGetTokenAddress(t *testing.T) {
	// USDT on Ethereum
	addr, ok := GetTokenAddress("USDT", 1)
	assert.True(t, ok)
	assert.Equal(t, "0xdAC17F958D2ee523a2206206994597C13D831ec7", addr.String())

	// USDC on Polygon
	addr, ok = GetTokenAddress("USDC", 137)
	assert.True(t, ok)
	assert.Equal(t, "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174", addr.String())

	// Unknown token
	_, ok = GetTokenAddress("UNKNOWN", 1)
	assert.False(t, ok)

	// Token on unsupported chain
	_, ok = GetTokenAddress("USDT", 999999)
	assert.False(t, ok)
}

func TestTransferData(t *testing.T) {
	// Create a mock token and verify transfer data encoding
	addr := mustDecodeHex("742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	var toAddr [20]byte
	copy(toAddr[:], addr)

	amount := big.NewInt(1000000)

	// Transfer function selector is 0xa9059cbb
	data := append(fnTransfer, padAddress(toAddr)...)
	data = append(data, padBigInt(amount)...)

	// Verify the data is correctly formatted
	assert.Equal(t, 68, len(data)) // 4 bytes selector + 32 bytes address + 32 bytes amount
	assert.Equal(t, byte(0xa9), data[0])
	assert.Equal(t, byte(0x05), data[1])
	assert.Equal(t, byte(0x9c), data[2])
	assert.Equal(t, byte(0xbb), data[3])
}
