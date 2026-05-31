package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khizar/bc-lib/internal/encoding"
	"github.com/khizar/bc-lib/pkg/types"
)

// TestSignLegacyEIP155Vector reproduces the canonical signed transaction from
// the EIP-155 specification, validating the entire signing pipeline:
// secp256k1, RFC 6979, EIP-155 v encoding, and RLP serialization.
func TestSignLegacyEIP155Vector(t *testing.T) {
	w, err := FromPrivateKeyHex("4646464646464646464646464646464646464646464646464646464646464646")
	require.NoError(t, err)

	to := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")
	tx := &types.Transaction{
		Type:     types.LegacyTxType,
		Nonce:    9,
		GasPrice: big.NewInt(20000000000),
		GasLimit: 21000,
		To:       &to,
		Value:    big.NewInt(1000000000000000000),
		Data:     []byte{},
		ChainID:  big.NewInt(1),
	}

	signed, err := w.SignTransaction(tx)
	require.NoError(t, err)

	expected := "f86c098504a817c800825208943535353535353535353535353535353535353535880de0b6b3a76400008025a028ef61340bd939bc2195fe537567866003e1a15d3c71ff63e1590620aa636276a067cbe9d8997f761aecb703304b3800ccf555c9f3dc64214b297fb1966a3b6d83"
	assert.Equal(t, expected, hex.EncodeToString(signed.RawTransaction))

	// v = 37 for chainID 1 and recovery id 0.
	assert.Equal(t, int64(37), signed.V.Int64())

	// The original transaction must be left unmodified.
	assert.Nil(t, tx.V)
}

func TestSignAndRecoverEIP1559(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	to := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	tx := &types.Transaction{
		Type:                 types.DynamicFeeTxType,
		ChainID:              big.NewInt(1),
		Nonce:                7,
		To:                   &to,
		Value:                big.NewInt(1000),
		Data:                 []byte{0xde, 0xad, 0xbe, 0xef},
		GasLimit:             50000,
		MaxFeePerGas:         big.NewInt(30000000000),
		MaxPriorityFeePerGas: big.NewInt(2000000000),
	}

	signed, err := w.SignTransaction(tx)
	require.NoError(t, err)
	assert.Equal(t, w.Address(), signed.From)

	// y-parity must be 0 or 1 for typed transactions.
	assert.True(t, signed.V.Int64() == 0 || signed.V.Int64() == 1)

	recovered, err := encoding.Sender(signed.Transaction)
	require.NoError(t, err)
	assert.Equal(t, w.Address(), recovered)
}

func TestSignLegacyRecover(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	to := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	tx := &types.Transaction{
		Type:     types.LegacyTxType,
		ChainID:  big.NewInt(11155111),
		Nonce:    3,
		To:       &to,
		Value:    big.NewInt(500),
		GasPrice: big.NewInt(15000000000),
		GasLimit: 21000,
	}

	signed, err := w.SignTransaction(tx)
	require.NoError(t, err)

	recovered, err := encoding.Sender(signed.Transaction)
	require.NoError(t, err)
	assert.Equal(t, w.Address(), recovered)
}

// storageKey returns a 32-byte hash whose last byte is v.
func storageKey(v byte) types.Hash {
	var h types.Hash
	h[31] = v
	return h
}

func TestSignAccessListTx(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	to := types.MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	accessed := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")
	tx := &types.Transaction{
		Type:     types.AccessListTxType,
		ChainID:  big.NewInt(1),
		Nonce:    2,
		To:       &to,
		Value:    big.NewInt(0),
		GasPrice: big.NewInt(10000000000),
		GasLimit: 60000,
		AccessList: types.AccessList{
			{
				Address:     accessed,
				StorageKeys: []types.Hash{storageKey(1)},
			},
		},
	}

	signed, err := w.SignTransaction(tx)
	require.NoError(t, err)
	require.NotNil(t, signed.RawTransaction)
	// Typed transaction envelope: first byte is the type identifier.
	assert.Equal(t, byte(types.AccessListTxType), signed.RawTransaction[0])

	recovered, err := encoding.Sender(signed.Transaction)
	require.NoError(t, err)
	assert.Equal(t, w.Address(), recovered)
}

func TestSignRequiresChainID(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	tx := &types.Transaction{
		Type:                 types.DynamicFeeTxType,
		Nonce:                1,
		Value:                big.NewInt(1),
		GasLimit:             21000,
		MaxFeePerGas:         big.NewInt(1),
		MaxPriorityFeePerGas: big.NewInt(1),
	}

	_, err = w.SignTransaction(tx)
	assert.Error(t, err)
}
