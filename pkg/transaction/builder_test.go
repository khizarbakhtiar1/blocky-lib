package transaction

import (
	"math/big"
	"testing"

	"github.com/khizar/bc-lib/pkg/chains"
	"github.com/khizar/bc-lib/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	require.NotNil(t, builder)
	require.NotNil(t, builder.tx)
	assert.Equal(t, types.DynamicFeeTxType, builder.tx.Type)
}

func TestNewLegacyBuilder(t *testing.T) {
	builder := NewLegacyBuilder()
	require.NotNil(t, builder)
	assert.Equal(t, types.LegacyTxType, builder.tx.Type)
}

func TestBuilderTo(t *testing.T) {
	addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	builder := NewBuilder().To(addr)

	require.NotNil(t, builder.tx.To)
	assert.Equal(t, addr, *builder.tx.To)
}

func TestBuilderValue(t *testing.T) {
	tests := []struct {
		name     string
		value    *big.Int
		expected *big.Int
	}{
		{
			name:     "positive value",
			value:    big.NewInt(1000000),
			expected: big.NewInt(1000000),
		},
		{
			name:     "zero value",
			value:    big.NewInt(0),
			expected: big.NewInt(0),
		},
		{
			name:     "nil value",
			value:    nil,
			expected: big.NewInt(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder().Value(tt.value)
			assert.Equal(t, 0, tt.expected.Cmp(builder.tx.Value))
		})
	}
}

func TestBuilderValueEther(t *testing.T) {
	builder := NewBuilder().ValueEther(1.5)

	expected, _ := types.ParseEther("1.5")
	assert.Equal(t, 0, expected.Cmp(builder.tx.Value))
}

func TestBuilderGasSettings(t *testing.T) {
	maxFee := big.NewInt(30000000000)
	priorityFee := big.NewInt(2000000000)

	builder := NewBuilder().
		MaxFeePerGas(maxFee).
		MaxPriorityFeePerGas(priorityFee).
		GasLimit(21000)

	assert.Equal(t, types.DynamicFeeTxType, builder.tx.Type)
	assert.Equal(t, 0, maxFee.Cmp(builder.tx.MaxFeePerGas))
	assert.Equal(t, 0, priorityFee.Cmp(builder.tx.MaxPriorityFeePerGas))
	assert.Equal(t, uint64(21000), builder.tx.GasLimit)
}

func TestBuilderLegacyGas(t *testing.T) {
	gasPrice := big.NewInt(20000000000)

	builder := NewBuilder().
		GasPrice(gasPrice).
		GasLimit(21000)

	assert.Equal(t, types.LegacyTxType, builder.tx.Type)
	assert.Equal(t, 0, gasPrice.Cmp(builder.tx.GasPrice))
}

func TestBuilderChainID(t *testing.T) {
	chainID := big.NewInt(1)

	builder := NewBuilder().ChainID(chainID)

	assert.Equal(t, 0, chainID.Cmp(builder.tx.ChainID))
}

func TestBuilderBuild(t *testing.T) {
	addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	tx, err := NewBuilder().
		To(addr).
		Value(big.NewInt(1000000)).
		ChainID(big.NewInt(1)).
		MaxFeePerGas(big.NewInt(30000000000)).
		MaxPriorityFeePerGas(big.NewInt(2000000000)).
		GasLimit(21000).
		Nonce(5).
		Build()

	require.NoError(t, err)
	require.NotNil(t, tx)
	assert.Equal(t, uint64(5), tx.Nonce)
	assert.Equal(t, uint64(21000), tx.GasLimit)
}

func TestBuilderBuildValidation(t *testing.T) {
	addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")

	tests := []struct {
		name        string
		buildFunc   func() (*types.Transaction, error)
		expectError bool
	}{
		{
			name: "missing chain ID",
			buildFunc: func() (*types.Transaction, error) {
				return NewBuilder().
					To(addr).
					MaxFeePerGas(big.NewInt(30000000000)).
					MaxPriorityFeePerGas(big.NewInt(2000000000)).
					Build()
			},
			expectError: true,
		},
		{
			name: "missing maxFeePerGas for EIP-1559",
			buildFunc: func() (*types.Transaction, error) {
				return NewBuilder().
					To(addr).
					ChainID(big.NewInt(1)).
					MaxPriorityFeePerGas(big.NewInt(2000000000)).
					Build()
			},
			expectError: true,
		},
		{
			name: "priority fee exceeds max fee",
			buildFunc: func() (*types.Transaction, error) {
				return NewBuilder().
					To(addr).
					ChainID(big.NewInt(1)).
					MaxFeePerGas(big.NewInt(1000000000)).
					MaxPriorityFeePerGas(big.NewInt(2000000000)).
					Build()
			},
			expectError: true,
		},
		{
			name: "missing gas price for legacy",
			buildFunc: func() (*types.Transaction, error) {
				return NewLegacyBuilder().
					To(addr).
					ChainID(big.NewInt(1)).
					Build()
			},
			expectError: true,
		},
		{
			name: "zero gas limit",
			buildFunc: func() (*types.Transaction, error) {
				return NewBuilder().
					To(addr).
					ChainID(big.NewInt(1)).
					MaxFeePerGas(big.NewInt(30000000000)).
					MaxPriorityFeePerGas(big.NewInt(2000000000)).
					GasLimit(0).
					Build()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx, err := tt.buildFunc()
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	value := big.NewInt(1000000000000000000) // 1 ETH

	tx, err := Transfer(addr, value, chains.EthereumMainnet).
		MaxFeePerGas(big.NewInt(30000000000)).
		MaxPriorityFeePerGas(big.NewInt(2000000000)).
		Build()

	require.NoError(t, err)
	assert.Equal(t, addr, *tx.To)
	assert.Equal(t, uint64(21000), tx.GasLimit)
	assert.Equal(t, 0, value.Cmp(tx.Value))
}

func TestContractCall(t *testing.T) {
	addr := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	data := []byte{0x12, 0x34, 0x56, 0x78}

	tx, err := ContractCall(addr, data, chains.EthereumMainnet).
		MaxFeePerGas(big.NewInt(30000000000)).
		MaxPriorityFeePerGas(big.NewInt(2000000000)).
		GasLimit(100000).
		Build()

	require.NoError(t, err)
	assert.Equal(t, addr, *tx.To)
	assert.Equal(t, data, tx.Data)
}

func TestContractDeploy(t *testing.T) {
	bytecode := []byte{0x60, 0x80, 0x60, 0x40}

	tx, err := ContractDeploy(bytecode, chains.EthereumMainnet).
		MaxFeePerGas(big.NewInt(30000000000)).
		MaxPriorityFeePerGas(big.NewInt(2000000000)).
		GasLimit(1000000).
		Build()

	require.NoError(t, err)
	assert.Nil(t, tx.To) // Contract creation has no 'to' address
	assert.Equal(t, bytecode, tx.Data)
}
