package crypto

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKnownAddressVectors checks that well-known private keys map to their
// canonical Ethereum addresses, proving the secp256k1 group law and address
// derivation are correct.
func TestKnownAddressVectors(t *testing.T) {
	vectors := []struct {
		privHex string
		address string
	}{
		{
			privHex: "0000000000000000000000000000000000000000000000000000000000000001",
			address: "0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf",
		},
		{
			privHex: "0000000000000000000000000000000000000000000000000000000000000002",
			address: "0x2B5AD5c4795c026514f8317c7a215E218DcCD6cF",
		},
		{
			// Vitalik's classic test key from the EIP-155 specification.
			privHex: "4646464646464646464646464646464646464646464646464646464646464646",
			address: "0x9d8A62f656a8d1615C1294fd71e9CFb3E4855A4F",
		},
	}

	for _, v := range vectors {
		priv, err := HexToECDSA(v.privHex)
		require.NoError(t, err)
		addr := PubkeyToAddress(&priv.PublicKey)
		assert.Equal(t, v.address, addr.String(), "priv %s", v.privHex)
	}
}

// TestEIP155SignatureVector reproduces the signature from the EIP-155
// specification, validating RFC 6979 deterministic nonce generation together
// with secp256k1 signing.
func TestEIP155SignatureVector(t *testing.T) {
	priv, err := HexToECDSA("4646464646464646464646464646464646464646464646464646464646464646")
	require.NoError(t, err)

	// keccak256 of the EIP-155 signing payload for the example transaction.
	hash, err := hex.DecodeString("daf5a779ae972f972197303d7b574746c7ef83eadac0f2791ad23db92e4c8e53")
	require.NoError(t, err)

	sig, err := Sign(hash, priv)
	require.NoError(t, err)

	r := new(big.Int).SetBytes(sig[0:32])
	s := new(big.Int).SetBytes(sig[32:64])
	recid := sig[64]

	expectedR, _ := new(big.Int).SetString("28ef61340bd939bc2195fe537567866003e1a15d3c71ff63e1590620aa636276", 16)
	expectedS, _ := new(big.Int).SetString("67cbe9d8997f761aecb703304b3800ccf555c9f3dc64214b297fb1966a3b6d83", 16)

	assert.Equal(t, 0, expectedR.Cmp(r), "r mismatch")
	assert.Equal(t, 0, expectedS.Cmp(s), "s mismatch")
	// EIP-155 v = 37 with chainID 1 implies recovery id 0 (v = recid + 35 + 2*chainID).
	assert.Equal(t, byte(0), recid, "recovery id mismatch")
}

func TestSignRecoverRoundTrip(t *testing.T) {
	priv, err := GenerateKey()
	require.NoError(t, err)

	hash := Keccak256([]byte("recover me"))
	sig, err := Sign(hash, priv)
	require.NoError(t, err)

	pub, err := SigToPub(hash, sig)
	require.NoError(t, err)
	assert.Equal(t, 0, priv.PublicKey.X.Cmp(pub.X))
	assert.Equal(t, 0, priv.PublicKey.Y.Cmp(pub.Y))

	recovered := PubkeyToAddress(pub)
	assert.Equal(t, PubkeyToAddress(&priv.PublicKey), recovered)
}

func TestLowSEnforced(t *testing.T) {
	priv, err := GenerateKey()
	require.NoError(t, err)

	halfN := new(big.Int).Rsh(S256().Params().N, 1)
	for i := 0; i < 20; i++ {
		hash := Keccak256([]byte{byte(i)})
		sig, err := Sign(hash, priv)
		require.NoError(t, err)
		s := new(big.Int).SetBytes(sig[32:64])
		assert.True(t, s.Cmp(halfN) <= 0, "s must be canonical (low-S)")
	}
}
