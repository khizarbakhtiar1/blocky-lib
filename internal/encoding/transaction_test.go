package encoding

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/khizar/bc-lib/pkg/types"
)

func TestEncodeAccessListEmpty(t *testing.T) {
	raw, err := encodeAccessList(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// An empty access list must encode as an empty RLP list (0xc0), not an
	// empty string (0x80).
	if !bytes.Equal(raw, []byte{0xc0}) {
		t.Fatalf("empty access list = %x, want c0", raw)
	}
}

func TestEncodeAccessListSingle(t *testing.T) {
	addr := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")
	var key types.Hash
	key[31] = 1

	raw, err := encodeAccessList(types.AccessList{{Address: addr, StorageKeys: []types.Hash{key}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must be a non-empty list whose prefix is in the list range (>= 0xc0).
	if len(raw) == 0 || raw[0] < 0xc0 {
		t.Fatalf("access list not encoded as list: %x", raw)
	}
}

func TestEncodeSignedRequiresSignature(t *testing.T) {
	to := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")
	tx := &types.Transaction{
		Type:     types.LegacyTxType,
		Nonce:    1,
		GasPrice: big.NewInt(1),
		GasLimit: 21000,
		To:       &to,
		Value:    big.NewInt(1),
		ChainID:  big.NewInt(1),
	}
	if _, err := EncodeSigned(tx); err == nil {
		t.Fatal("expected error for unsigned transaction")
	}
}

func TestSigningHashDiffersByType(t *testing.T) {
	to := types.MustAddressFromHex("0x3535353535353535353535353535353535353535")
	base := types.Transaction{
		Nonce:    1,
		GasLimit: 21000,
		To:       &to,
		Value:    big.NewInt(1),
		ChainID:  big.NewInt(1),
		GasPrice: big.NewInt(1),

		MaxFeePerGas:         big.NewInt(2),
		MaxPriorityFeePerGas: big.NewInt(1),
	}

	legacy := base
	legacy.Type = types.LegacyTxType
	dynamic := base
	dynamic.Type = types.DynamicFeeTxType

	h1, err := SigningHash(&legacy)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := SigningHash(&dynamic)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Fatal("legacy and dynamic-fee signing hashes must differ")
	}
}
