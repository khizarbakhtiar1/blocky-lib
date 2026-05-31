package crypto

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS256(t *testing.T) {
	curve := S256()
	require.NotNil(t, curve)

	params := curve.Params()
	assert.Equal(t, 256, params.BitSize)
	assert.Equal(t, "secp256k1", params.Name)
	assert.Equal(t, big.NewInt(7), params.B)

	// The generator point must lie on the curve.
	assert.True(t, curve.IsOnCurve(params.Gx, params.Gy))

	// N is the documented secp256k1 group order.
	expectedN, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
	assert.Equal(t, 0, expectedN.Cmp(params.N))
}

func TestFromECDSA(t *testing.T) {
	// nil key
	assert.Nil(t, FromECDSA(nil))

	// valid key
	key, err := GenerateKey()
	require.NoError(t, err)
	b := FromECDSA(key)
	assert.NotNil(t, b)
	assert.True(t, len(b) > 0)
}

func TestToECDSA(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	// Round-trip: export then import
	exported := FromECDSA(key)
	imported, err := ToECDSA(exported)
	require.NoError(t, err)
	assert.Equal(t, key.D.Bytes(), imported.D.Bytes())

	// Empty bytes should fail
	_, err = ToECDSA([]byte{})
	assert.Error(t, err)
}

func TestHexToECDSA(t *testing.T) {
	// Generate a key, export to hex, re-import via ToECDSA
	key, err := GenerateKey()
	require.NoError(t, err)

	exported := FromECDSA(key)
	reimported, err := ToECDSA(exported)
	require.NoError(t, err)
	assert.Equal(t, key.D.Bytes(), reimported.D.Bytes())

	// Invalid hex
	_, err = HexToECDSA("not-valid-hex")
	assert.Error(t, err)

	// Empty hex
	_, err = HexToECDSA("")
	assert.Error(t, err)

	// With 0x prefix only - should fail
	_, err = HexToECDSA("0x")
	assert.Error(t, err)
}
