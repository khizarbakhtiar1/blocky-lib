package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToWei(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		unit     Unit
		expected string
	}{
		{
			name:     "1 ether to wei",
			value:    big.NewInt(1),
			unit:     Ether,
			expected: "1000000000000000000",
		},
		{
			name:     "1 gwei to wei",
			value:    big.NewInt(1),
			unit:     Gwei,
			expected: "1000000000",
		},
		{
			name:     "1 wei to wei",
			value:    big.NewInt(1),
			unit:     Wei,
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToWei(tt.value, tt.unit)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestFromWei(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		unit     Unit
		expected string
	}{
		{
			name:     "1 ether from wei",
			value:    "1000000000000000000",
			unit:     Ether,
			expected: "1",
		},
		{
			name:     "1 gwei from wei",
			value:    "1000000000",
			unit:     Gwei,
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, _ := new(big.Int).SetString(tt.value, 10)
			result := FromWei(value, tt.unit)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

func TestEtherToWei(t *testing.T) {
	ether := big.NewInt(1)
	wei := EtherToWei(ether)
	assert.Equal(t, "1000000000000000000", wei.String())
}

func TestWeiToEther(t *testing.T) {
	wei, _ := new(big.Int).SetString("1000000000000000000", 10)
	ether := WeiToEther(wei)
	assert.Equal(t, "1", ether.String())
}

func TestGweiToWei(t *testing.T) {
	gwei := big.NewInt(100)
	wei := GweiToWei(gwei)
	assert.Equal(t, "100000000000", wei.String())
}

func TestWeiToGwei(t *testing.T) {
	wei, _ := new(big.Int).SetString("100000000000", 10)
	gwei := WeiToGwei(wei)
	assert.Equal(t, "100", gwei.String())
}

func TestParseEther(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "parse 1 ether",
			input:    "1",
			expected: "1000000000000000000",
			wantErr:  false,
		},
		{
			name:     "parse 0.5 ether",
			input:    "0.5",
			expected: "500000000000000000",
			wantErr:  false,
		},
		{
			name:    "invalid input",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseEther(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result.String())
			}
		})
	}
}

func TestFormatEther(t *testing.T) {
	wei, _ := new(big.Int).SetString("1000000000000000000", 10)
	formatted := FormatEther(wei)
	assert.Contains(t, formatted, "1")
}

func TestFormatGwei(t *testing.T) {
	wei, _ := new(big.Int).SetString("1000000000", 10)
	formatted := FormatGwei(wei)
	assert.Contains(t, formatted, "1")
}

func TestFormatEtherNil(t *testing.T) {
	formatted := FormatEther(nil)
	assert.Equal(t, "0", formatted)
}

func TestFormatGweiNil(t *testing.T) {
	formatted := FormatGwei(nil)
	assert.Equal(t, "0", formatted)
}
