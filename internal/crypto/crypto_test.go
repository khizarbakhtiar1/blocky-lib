package crypto

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeccak256(t *testing.T) {
	hash := Keccak256([]byte{})
	assert.Equal(t, 32, len(hash))

	hash1 := Keccak256([]byte("hello"))
	hash2 := Keccak256([]byte("hello"))
	assert.True(t, bytes.Equal(hash1, hash2))

	hash3 := Keccak256([]byte("world"))
	assert.False(t, bytes.Equal(hash1, hash3))

	hashConcat := Keccak256([]byte("hel"), []byte("lo"))
	assert.True(t, bytes.Equal(hash1, hashConcat))
}

func TestKeccak256Hash(t *testing.T) {
	h := Keccak256Hash([]byte("test"))
	assert.False(t, h.IsZero())
	assert.Equal(t, 32, len(h.Bytes()))
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)
	require.NotNil(t, key)
	assert.NotNil(t, key.PublicKey.X)
	assert.NotNil(t, key.PublicKey.Y)
	assert.NotNil(t, key.D)

	key2, err := GenerateKey()
	require.NoError(t, err)
	assert.NotEqual(t, key.D, key2.D)
}

func TestPubkeyToAddress(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	addr := PubkeyToAddress(&key.PublicKey)
	assert.False(t, addr.IsZero())
	assert.Equal(t, 20, len(addr.Bytes()))

	addr2 := PubkeyToAddress(&key.PublicKey)
	assert.Equal(t, addr, addr2)
}

func TestSign(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	hash := Keccak256([]byte("test message"))
	sig, err := Sign(hash, key)
	require.NoError(t, err)
	assert.Equal(t, 65, len(sig))

	_, err = Sign([]byte("short"), key)
	assert.Error(t, err)
}

func TestSignAndVerify(t *testing.T) {
	key, err := GenerateKey()
	require.NoError(t, err)

	hash := Keccak256([]byte("test message"))
	sig, err := Sign(hash, key)
	require.NoError(t, err)

	xBytes := make([]byte, 32)
	yBytes := make([]byte, 32)
	xB := key.PublicKey.X.Bytes()
	yB := key.PublicKey.Y.Bytes()
	copy(xBytes[32-len(xB):], xB)
	copy(yBytes[32-len(yB):], yB)
	pubBytes := append(xBytes, yBytes...)

	assert.True(t, VerifySignature(pubBytes, hash, sig))

	badHash := make([]byte, 32)
	copy(badHash, hash)
	badHash[0] ^= 0xff
	assert.False(t, VerifySignature(pubBytes, badHash, sig))
}

func TestVerifySignatureInvalidInputs(t *testing.T) {
	assert.False(t, VerifySignature(nil, nil, nil))
	assert.False(t, VerifySignature(make([]byte, 64), make([]byte, 32), make([]byte, 10)))
	assert.False(t, VerifySignature(make([]byte, 10), make([]byte, 32), make([]byte, 65)))
	assert.False(t, VerifySignature(make([]byte, 64), make([]byte, 10), make([]byte, 65)))
}
