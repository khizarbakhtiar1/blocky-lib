package encoding

import (
	"bytes"
	"fmt"
	"math/big"
)

// RLP encoding implementation for Ethereum transactions
// Based on the Recursive Length Prefix encoding scheme

// Raw represents an already RLP-encoded node. When passed to Encode it is
// emitted verbatim, which is useful for embedding pre-encoded structures such
// as EIP-2930 access lists inside a larger list.
type Raw []byte

// Encode encodes a value using RLP
func Encode(val interface{}) ([]byte, error) {
	switch v := val.(type) {
	case Raw:
		return []byte(v), nil
	case []byte:
		return encodeBytes(v), nil
	case string:
		return encodeBytes([]byte(v)), nil
	case uint64:
		return encodeUint64(v), nil
	case *big.Int:
		return encodeBigInt(v), nil
	case []interface{}:
		return encodeList(v)
	default:
		return nil, fmt.Errorf("unsupported type for RLP encoding: %T", val)
	}
}

// EncodeList encodes a list of items
func EncodeList(items ...interface{}) ([]byte, error) {
	return encodeList(items)
}

// encodeBytes encodes a byte slice
func encodeBytes(b []byte) []byte {
	if len(b) == 1 && b[0] < 0x80 {
		return b
	}
	return append(encodeLength(len(b), 0x80), b...)
}

// encodeUint64 encodes a uint64
func encodeUint64(i uint64) []byte {
	if i == 0 {
		return []byte{0x80}
	}
	return encodeBytes(bigIntToBytes(new(big.Int).SetUint64(i)))
}

// encodeBigInt encodes a big.Int
func encodeBigInt(i *big.Int) []byte {
	if i == nil || i.Sign() == 0 {
		return []byte{0x80}
	}
	return encodeBytes(bigIntToBytes(i))
}

// bigIntToBytes converts a big.Int to bytes without leading zeros
func bigIntToBytes(i *big.Int) []byte {
	if i.Sign() == 0 {
		return nil
	}
	return i.Bytes()
}

// encodeList encodes a list of items
func encodeList(items []interface{}) ([]byte, error) {
	var buf bytes.Buffer

	for _, item := range items {
		encoded, err := Encode(item)
		if err != nil {
			return nil, err
		}
		buf.Write(encoded)
	}

	content := buf.Bytes()
	return append(encodeLength(len(content), 0xc0), content...), nil
}

// encodeLength encodes the length prefix
func encodeLength(length int, offset byte) []byte {
	if length < 56 {
		return []byte{offset + byte(length)}
	}

	// For lengths >= 56, we need to encode the length of the length
	lengthBytes := bigIntToBytes(big.NewInt(int64(length)))
	return append([]byte{offset + 55 + byte(len(lengthBytes))}, lengthBytes...)
}

// Decode decodes RLP data
func Decode(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty RLP data")
	}

	prefix := data[0]

	switch {
	case prefix < 0x80:
		// Single byte
		return []byte{prefix}, 1, nil

	case prefix < 0xb8:
		// Short string (0-55 bytes)
		length := int(prefix - 0x80)
		if len(data) < 1+length {
			return nil, 0, fmt.Errorf("insufficient data for string")
		}
		return data[1 : 1+length], 1 + length, nil

	case prefix < 0xc0:
		// Long string (>55 bytes)
		lengthOfLength := int(prefix - 0xb7)
		if len(data) < 1+lengthOfLength {
			return nil, 0, fmt.Errorf("insufficient data for string length")
		}
		length := bytesToInt(data[1 : 1+lengthOfLength])
		if len(data) < 1+lengthOfLength+length {
			return nil, 0, fmt.Errorf("insufficient data for long string")
		}
		return data[1+lengthOfLength : 1+lengthOfLength+length], 1 + lengthOfLength + length, nil

	case prefix < 0xf8:
		// Short list (0-55 bytes total)
		length := int(prefix - 0xc0)
		if len(data) < 1+length {
			return nil, 0, fmt.Errorf("insufficient data for list")
		}
		return decodeList(data[1 : 1+length])

	default:
		// Long list (>55 bytes total)
		lengthOfLength := int(prefix - 0xf7)
		if len(data) < 1+lengthOfLength {
			return nil, 0, fmt.Errorf("insufficient data for list length")
		}
		length := bytesToInt(data[1 : 1+lengthOfLength])
		if len(data) < 1+lengthOfLength+length {
			return nil, 0, fmt.Errorf("insufficient data for long list")
		}
		return decodeList(data[1+lengthOfLength : 1+lengthOfLength+length])
	}
}

// decodeList decodes a list from RLP content
func decodeList(content []byte) ([]interface{}, int, error) {
	var items []interface{}
	offset := 0

	for offset < len(content) {
		item, consumed, err := Decode(content[offset:])
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
		offset += consumed
	}

	return items, len(content), nil
}

// bytesToInt converts bytes to int
func bytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}
