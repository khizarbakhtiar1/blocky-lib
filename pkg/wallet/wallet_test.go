package wallet

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khizar/bc-lib/internal/crypto"
)

func TestNewWallet(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)
	require.NotNil(t, w)

	assert.False(t, w.Address().IsZero())
	assert.NotNil(t, w.PrivateKey())
	assert.NotNil(t, w.PublicKey())

	w2, err := NewWallet()
	require.NoError(t, err)
	assert.NotEqual(t, w.Address(), w2.Address())
}

func TestFromPrivateKey(t *testing.T) {
	key, err := crypto.GenerateKey()
	require.NoError(t, err)

	w := FromPrivateKey(key)
	require.NotNil(t, w)
	assert.Equal(t, key, w.PrivateKey())
	assert.False(t, w.Address().IsZero())
}

func TestFromPrivateKeyHex(t *testing.T) {
	w1, err := NewWallet()
	require.NoError(t, err)

	hexKey := w1.PrivateKeyHex()
	assert.NotEmpty(t, hexKey)

	w2, err := FromPrivateKeyHex(hexKey)
	require.NoError(t, err)
	assert.Equal(t, w1.Address(), w2.Address())

	w3, err := FromPrivateKeyHex("0x" + hexKey)
	require.NoError(t, err)
	assert.Equal(t, w1.Address(), w3.Address())
}

func TestFromPrivateKeyHexInvalid(t *testing.T) {
	_, err := FromPrivateKeyHex("not-hex")
	assert.Error(t, err)

	_, err = FromPrivateKeyHex("")
	assert.Error(t, err)
}

func TestWalletPrivateKeyHex(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	hexKey := w.PrivateKeyHex()
	_, err = hex.DecodeString(hexKey)
	assert.NoError(t, err)
}

func TestWalletSign(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	hash := crypto.Keccak256([]byte("test"))
	sig, err := w.Sign(hash)
	require.NoError(t, err)
	assert.Equal(t, 65, len(sig))
}

func TestWalletSignTransaction(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	_, err = w.SignTransaction(nil)
	assert.Error(t, err)
}

func TestWalletString(t *testing.T) {
	w, err := NewWallet()
	require.NoError(t, err)

	str := w.String()
	assert.Contains(t, str, "0x")
	assert.Equal(t, w.Address().String(), str)
}
