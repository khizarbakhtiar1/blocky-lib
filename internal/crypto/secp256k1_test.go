package crypto

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestS256(t *testing.T) {
	curve := S256()
	assert.NotNil(t, curve)
	assert.IsType(t, elliptic.P256(), curve)
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
