package encoding

import (
	"fmt"
	"math/big"

	"github.com/khizar/bc-lib/internal/crypto"
	"github.com/khizar/bc-lib/pkg/types"
)

// toField returns the RLP representation of a transaction recipient: the 20
// address bytes, or an empty string for contract creation (nil To).
func toField(tx *types.Transaction) []byte {
	if tx.To == nil {
		return []byte{}
	}
	return tx.To[:]
}

// encodeAccessList RLP-encodes an EIP-2930 access list as a list of
// [address, [storageKeys...]] tuples.
func encodeAccessList(list types.AccessList) (Raw, error) {
	items := make([]interface{}, 0, len(list))
	for _, tuple := range list {
		keys := make([]interface{}, 0, len(tuple.StorageKeys))
		for _, k := range tuple.StorageKeys {
			key := k
			keys = append(keys, key[:])
		}
		keysEncoded, err := EncodeList(keys...)
		if err != nil {
			return nil, err
		}
		entry, err := EncodeList(tuple.Address[:], Raw(keysEncoded))
		if err != nil {
			return nil, err
		}
		items = append(items, Raw(entry))
	}
	encoded, err := EncodeList(items...)
	if err != nil {
		return nil, err
	}
	return Raw(encoded), nil
}

// SigningHash returns the Keccak-256 hash that must be signed for the given
// transaction, according to its type (legacy/EIP-155, EIP-2930 or EIP-1559).
func SigningHash(tx *types.Transaction) (types.Hash, error) {
	payload, err := signingPayload(tx)
	if err != nil {
		return types.Hash{}, err
	}
	return crypto.Keccak256Hash(payload), nil
}

func signingPayload(tx *types.Transaction) ([]byte, error) {
	switch tx.Type {
	case types.LegacyTxType:
		items := []interface{}{
			tx.Nonce,
			normalizeBig(tx.GasPrice),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
		}
		// EIP-155: append chainID, 0, 0 when a chain ID is present.
		if tx.ChainID != nil && tx.ChainID.Sign() > 0 {
			items = append(items, tx.ChainID, uint64(0), uint64(0))
		}
		return EncodeList(items...)

	case types.AccessListTxType:
		al, err := encodeAccessList(tx.AccessList)
		if err != nil {
			return nil, err
		}
		body, err := EncodeList(
			tx.ChainID,
			tx.Nonce,
			normalizeBig(tx.GasPrice),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
			al,
		)
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(types.AccessListTxType)}, body...), nil

	case types.DynamicFeeTxType:
		al, err := encodeAccessList(tx.AccessList)
		if err != nil {
			return nil, err
		}
		body, err := EncodeList(
			tx.ChainID,
			tx.Nonce,
			normalizeBig(tx.MaxPriorityFeePerGas),
			normalizeBig(tx.MaxFeePerGas),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
			al,
		)
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(types.DynamicFeeTxType)}, body...), nil

	default:
		return nil, fmt.Errorf("unsupported transaction type: %d", tx.Type)
	}
}

// EncodeSigned RLP-encodes a fully signed transaction (V, R, S populated) into
// the raw bytes that can be broadcast via eth_sendRawTransaction.
func EncodeSigned(tx *types.Transaction) ([]byte, error) {
	if tx.R == nil || tx.S == nil || tx.V == nil {
		return nil, fmt.Errorf("transaction is not signed")
	}

	switch tx.Type {
	case types.LegacyTxType:
		return EncodeList(
			tx.Nonce,
			normalizeBig(tx.GasPrice),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
			tx.V,
			tx.R,
			tx.S,
		)

	case types.AccessListTxType:
		al, err := encodeAccessList(tx.AccessList)
		if err != nil {
			return nil, err
		}
		body, err := EncodeList(
			tx.ChainID,
			tx.Nonce,
			normalizeBig(tx.GasPrice),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
			al,
			tx.V,
			tx.R,
			tx.S,
		)
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(types.AccessListTxType)}, body...), nil

	case types.DynamicFeeTxType:
		al, err := encodeAccessList(tx.AccessList)
		if err != nil {
			return nil, err
		}
		body, err := EncodeList(
			tx.ChainID,
			tx.Nonce,
			normalizeBig(tx.MaxPriorityFeePerGas),
			normalizeBig(tx.MaxFeePerGas),
			tx.GasLimit,
			toField(tx),
			normalizeBig(tx.Value),
			tx.Data,
			al,
			tx.V,
			tx.R,
			tx.S,
		)
		if err != nil {
			return nil, err
		}
		return append([]byte{byte(types.DynamicFeeTxType)}, body...), nil

	default:
		return nil, fmt.Errorf("unsupported transaction type: %d", tx.Type)
	}
}

// TxHash returns the transaction hash (Keccak-256 of the signed encoding).
func TxHash(rawSigned []byte) types.Hash {
	return crypto.Keccak256Hash(rawSigned)
}

// Sender recovers the address that signed the transaction from its signature.
func Sender(tx *types.Transaction) (types.Address, error) {
	if tx.R == nil || tx.S == nil || tx.V == nil {
		return types.Address{}, fmt.Errorf("transaction is not signed")
	}

	var recid byte
	switch tx.Type {
	case types.LegacyTxType:
		if tx.ChainID != nil && tx.ChainID.Sign() > 0 {
			// recid = v - 35 - 2*chainID (EIP-155)
			r := new(big.Int).Sub(tx.V, big.NewInt(35))
			r.Sub(r, new(big.Int).Mul(tx.ChainID, big.NewInt(2)))
			recid = byte(r.Int64())
		} else {
			// recid = v - 27 (pre-EIP-155)
			recid = byte(new(big.Int).Sub(tx.V, big.NewInt(27)).Int64())
		}
	default:
		recid = byte(tx.V.Int64())
	}

	hash, err := SigningHash(tx)
	if err != nil {
		return types.Address{}, err
	}

	sig := make([]byte, 65)
	copy(sig[0:32], leftPad32(tx.R))
	copy(sig[32:64], leftPad32(tx.S))
	sig[64] = recid

	pub, err := crypto.SigToPub(hash[:], sig)
	if err != nil {
		return types.Address{}, err
	}
	return crypto.PubkeyToAddress(pub), nil
}

// leftPad32 returns the big-endian encoding of n left-padded to 32 bytes.
func leftPad32(n *big.Int) []byte {
	b := n.Bytes()
	if len(b) >= 32 {
		return b
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

// normalizeBig returns a non-nil big.Int so RLP encodes a missing value as 0.
func normalizeBig(v *big.Int) *big.Int {
	if v == nil {
		return big.NewInt(0)
	}
	return v
}
