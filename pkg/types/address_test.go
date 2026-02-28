package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddressFromHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid address with 0x prefix",
			input:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			wantErr: false,
		},
		{
			name:    "valid address without 0x prefix",
			input:   "742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",
			wantErr: false,
		},
		{
			name:    "invalid hex characters",
			input:   "0xGGGGGGGG",
			wantErr: true,
		},
		{
			name:    "too short",
			input:   "0x742d35Cc",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := AddressFromHex(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, Address{}, addr)
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	addr := MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	str := addr.String()
	
	assert.True(t, len(str) > 0)
	assert.Contains(t, str, "0x")
}

func TestAddressIsZero(t *testing.T) {
	zero := ZeroAddress()
	assert.True(t, zero.IsZero())
	
	nonZero := MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	assert.False(t, nonZero.IsZero())
}

func TestAddressBytes(t *testing.T) {
	addr := MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	bytes := addr.Bytes()
	
	assert.Equal(t, 20, len(bytes))
}

func TestHashFromHex(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid hash with 0x prefix",
			input:   "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "valid hash without 0x prefix",
			input:   "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			wantErr: false,
		},
		{
			name:    "invalid length",
			input:   "0x1234",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashFromHex(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, Hash{}, hash)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	hash := MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	str := hash.String()
	
	assert.True(t, len(str) > 0)
	assert.Contains(t, str, "0x")
}

func TestHashIsZero(t *testing.T) {
	zero := ZeroHash()
	assert.True(t, zero.IsZero())
	
	nonZero := MustHashFromHex("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
	assert.False(t, nonZero.IsZero())
}

func TestMustAddressFromHex(t *testing.T) {
	// Should not panic with valid address
	assert.NotPanics(t, func() {
		MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	})
	
	// Should panic with invalid address
	assert.Panics(t, func() {
		MustAddressFromHex("invalid")
	})
}

func TestAddressValidate(t *testing.T) {
	zero := ZeroAddress()
	err := zero.Validate()
	assert.Error(t, err)
	
	valid := MustAddressFromHex("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0")
	err = valid.Validate()
	assert.NoError(t, err)
}

func TestNewAddress(t *testing.T) {
	// Valid 20 bytes
	bytes := make([]byte, 20)
	addr, err := NewAddress(bytes)
	require.NoError(t, err)
	assert.Equal(t, 20, len(addr))
	
	// Invalid length
	invalidBytes := make([]byte, 10)
	_, err = NewAddress(invalidBytes)
	assert.Error(t, err)
}

func TestNewHash(t *testing.T) {
	// Valid 32 bytes
	bytes := make([]byte, 32)
	hash, err := NewHash(bytes)
	require.NoError(t, err)
	assert.Equal(t, 32, len(hash))
	
	// Invalid length
	invalidBytes := make([]byte, 10)
	_, err = NewHash(invalidBytes)
	assert.Error(t, err)
}

