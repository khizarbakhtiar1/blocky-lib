package encoding

import (
	"bytes"
	"math/big"
	"testing"
)

func TestEncodeBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "single byte below 0x80",
			input:    []byte{0x00},
			expected: []byte{0x00},
		},
		{
			name:     "single byte 0x7f",
			input:    []byte{0x7f},
			expected: []byte{0x7f},
		},
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: []byte{0x80},
		},
		{
			name:     "short string",
			input:    []byte("dog"),
			expected: []byte{0x83, 'd', 'o', 'g'},
		},
		{
			name:     "55 byte string",
			input:    bytes.Repeat([]byte{'a'}, 55),
			expected: append([]byte{0x80 + 55}, bytes.Repeat([]byte{'a'}, 55)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    uint64
		expected []byte
	}{
		{
			name:     "zero",
			input:    0,
			expected: []byte{0x80},
		},
		{
			name:     "one",
			input:    1,
			expected: []byte{0x01},
		},
		{
			name:     "127",
			input:    127,
			expected: []byte{0x7f},
		},
		{
			name:     "128",
			input:    128,
			expected: []byte{0x81, 0x80},
		},
		{
			name:     "1024",
			input:    1024,
			expected: []byte{0x82, 0x04, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeBigInt(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected []byte
	}{
		{
			name:     "nil",
			input:    nil,
			expected: []byte{0x80},
		},
		{
			name:     "zero",
			input:    big.NewInt(0),
			expected: []byte{0x80},
		},
		{
			name:     "one",
			input:    big.NewInt(1),
			expected: []byte{0x01},
		},
		{
			name:     "large number",
			input:    big.NewInt(0x0400),
			expected: []byte{0x82, 0x04, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Encode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEncodeList(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []byte
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: []byte{0xc0},
		},
		{
			name:     "cat and dog",
			input:    []interface{}{[]byte("cat"), []byte("dog")},
			expected: []byte{0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g'},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeList(tt.input...)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected interface{}
	}{
		{
			name:     "single byte",
			input:    []byte{0x7f},
			expected: []byte{0x7f},
		},
		{
			name:     "empty string",
			input:    []byte{0x80},
			expected: []byte{},
		},
		{
			name:     "short string",
			input:    []byte{0x83, 'd', 'o', 'g'},
			expected: []byte("dog"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := Decode(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if expectedBytes, ok := tt.expected.([]byte); ok {
				if resultBytes, ok := result.([]byte); ok {
					if !bytes.Equal(resultBytes, expectedBytes) {
						t.Errorf("got %v, want %v", result, tt.expected)
					}
				}
			}
		})
	}
}
