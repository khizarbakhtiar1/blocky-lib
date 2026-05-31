package crypto

import (
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"

	"github.com/khizar/bc-lib/pkg/types"
)

// Keccak256 computes the Keccak256 hash of the input.
func Keccak256(data ...[]byte) []byte {
	hasher := sha3.NewLegacyKeccak256()
	for _, b := range data {
		hasher.Write(b)
	}
	return hasher.Sum(nil)
}

// Keccak256Hash computes the Keccak256 hash and returns a Hash type.
func Keccak256Hash(data ...[]byte) types.Hash {
	hash := Keccak256(data...)
	h, _ := types.NewHash(hash)
	return h
}

// PubkeyToAddress converts an ECDSA public key to an Ethereum address.
func PubkeyToAddress(pubkey *ecdsa.PublicKey) types.Address {
	// Serialize the public key (uncompressed, without the 0x04 prefix).
	// Each coordinate must be exactly 32 bytes, left-padded with zeros.
	pubBytes := append(paddedBytes(pubkey.X, 32), paddedBytes(pubkey.Y, 32)...)

	// The address is the last 20 bytes of the Keccak-256 hash of the key.
	hash := Keccak256(pubBytes)

	var addr types.Address
	copy(addr[:], hash[12:])
	return addr
}

// GenerateKey generates a new secp256k1 private key using a CSPRNG.
func GenerateKey() (*ecdsa.PrivateKey, error) {
	curve := S256()
	n := curve.Params().N

	// Reject samples outside [1, N-1] to avoid modulo bias.
	one := big.NewInt(1)
	max := new(big.Int).Sub(n, one)
	for {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return nil, err
		}
		d := new(big.Int).SetBytes(b)
		if d.Cmp(one) < 0 || d.Cmp(max) > 0 {
			continue
		}

		priv := new(ecdsa.PrivateKey)
		priv.PublicKey.Curve = curve
		priv.D = d
		priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())
		return priv, nil
	}
}

// Sign signs a 32-byte hash with the private key and returns a 65-byte
// signature in [R || S || V] format, where V is the recovery id (0 or 1).
// The signature uses a deterministic nonce (RFC 6979) and a normalized
// low-S value, matching Ethereum's signature requirements (EIP-2).
func Sign(hash []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	if privateKey == nil || privateKey.D == nil {
		return nil, fmt.Errorf("nil private key")
	}

	curve := S256()
	n := curve.Params().N
	d := privateKey.D
	e := new(big.Int).SetBytes(hash)

	halfN := new(big.Int).Rsh(n, 1)

	for i := 0; ; i++ {
		k := rfc6979Nonce(d, hash, n, i)

		// (x1, y1) = k*G
		x1, y1 := curve.ScalarBaseMult(k.Bytes())
		r := new(big.Int).Mod(x1, n)
		if r.Sign() == 0 {
			continue
		}

		// s = k^-1 * (e + r*d) mod n
		kInv := new(big.Int).ModInverse(k, n)
		s := new(big.Int).Mul(r, d)
		s.Add(s, e)
		s.Mul(s, kInv)
		s.Mod(s, n)
		if s.Sign() == 0 {
			continue
		}

		// Recovery id encodes the parity of y1 and whether x1 overflowed N.
		recid := byte(y1.Bit(0))
		if x1.Cmp(n) >= 0 {
			recid |= 2
		}

		// Enforce low-S (EIP-2). Flipping s also flips the y parity.
		if s.Cmp(halfN) > 0 {
			s.Sub(n, s)
			recid ^= 1
		}

		sig := make([]byte, 65)
		copy(sig[0:32], paddedBytes(r, 32))
		copy(sig[32:64], paddedBytes(s, 32))
		sig[64] = recid
		return sig, nil
	}
}

// VerifySignature verifies an ECDSA signature against a public key and hash.
// pubkey must be a 64-byte uncompressed public key (X || Y, without the 0x04
// prefix). signature must be 65 bytes in [R || S || V] format; the V byte is
// ignored during verification.
func VerifySignature(pubkey, hash, signature []byte) bool {
	if len(signature) != 65 || len(pubkey) != 64 || len(hash) != 32 {
		return false
	}

	curve := S256()
	n := curve.Params().N

	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])
	if r.Sign() <= 0 || s.Sign() <= 0 || r.Cmp(n) >= 0 || s.Cmp(n) >= 0 {
		return false
	}

	x := new(big.Int).SetBytes(pubkey[0:32])
	y := new(big.Int).SetBytes(pubkey[32:64])
	if !curve.IsOnCurve(x, y) {
		return false
	}

	e := new(big.Int).SetBytes(hash)

	// w = s^-1 mod n; u1 = e*w; u2 = r*w
	w := new(big.Int).ModInverse(s, n)
	u1 := new(big.Int).Mul(e, w)
	u1.Mod(u1, n)
	u2 := new(big.Int).Mul(r, w)
	u2.Mod(u2, n)

	x1, y1 := curve.ScalarBaseMult(u1.Bytes())
	x2, y2 := curve.ScalarMult(x, y, u2.Bytes())
	px, _ := curve.Add(x1, y1, x2, y2)
	if px.Sign() == 0 && y2.Sign() == 0 {
		return false
	}

	px.Mod(px, n)
	return px.Cmp(r) == 0
}

// Ecrecover recovers the 65-byte uncompressed public key (0x04 || X || Y) that
// produced the signature over hash. The signature must be 65 bytes in
// [R || S || V] format where V is the recovery id (0..3).
func Ecrecover(hash, signature []byte) ([]byte, error) {
	pub, err := SigToPub(hash, signature)
	if err != nil {
		return nil, err
	}
	out := make([]byte, 65)
	out[0] = 0x04
	copy(out[1:33], paddedBytes(pub.X, 32))
	copy(out[33:65], paddedBytes(pub.Y, 32))
	return out, nil
}

// SigToPub recovers the public key that produced the signature over hash.
func SigToPub(hash, signature []byte) (*ecdsa.PublicKey, error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash must be 32 bytes, got %d", len(hash))
	}
	if len(signature) != 65 {
		return nil, fmt.Errorf("signature must be 65 bytes, got %d", len(signature))
	}

	curve := S256()
	p := curve.Params().P
	n := curve.Params().N

	r := new(big.Int).SetBytes(signature[0:32])
	s := new(big.Int).SetBytes(signature[32:64])
	recid := signature[64]
	if recid > 3 {
		return nil, fmt.Errorf("invalid recovery id: %d", recid)
	}
	if r.Sign() <= 0 || s.Sign() <= 0 || r.Cmp(n) >= 0 || s.Cmp(n) >= 0 {
		return nil, fmt.Errorf("signature values out of range")
	}

	// Reconstruct R's x-coordinate, adding N if the recovery id indicates an
	// overflow (recid >= 2).
	x := new(big.Int).Set(r)
	if recid >= 2 {
		x.Add(x, n)
		if x.Cmp(p) >= 0 {
			return nil, fmt.Errorf("recovered x out of range")
		}
	}

	// y^2 = x^3 + 7 mod p; pick the root with the requested parity.
	ySq := new(big.Int).Mul(x, x)
	ySq.Mul(ySq, x)
	ySq.Add(ySq, curve.Params().B)
	ySq.Mod(ySq, p)

	y := new(big.Int).ModSqrt(ySq, p)
	if y == nil {
		return nil, fmt.Errorf("invalid signature: R is not on the curve")
	}
	if y.Bit(0) != uint(recid&1) {
		y.Sub(p, y)
	}

	if !curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("recovered point is not on the curve")
	}

	// Q = r^-1 * (s*R - e*G)
	e := new(big.Int).SetBytes(hash)
	rInv := new(big.Int).ModInverse(r, n)

	// s*R
	sRx, sRy := curve.ScalarMult(x, y, s.Bytes())
	// e*G, then negate to subtract.
	eGx, eGy := curve.ScalarBaseMult(new(big.Int).Mod(e, n).Bytes())
	negEGy := new(big.Int).Sub(p, eGy)
	sumX, sumY := curve.Add(sRx, sRy, eGx, negEGy)

	qx, qy := curve.ScalarMult(sumX, sumY, rInv.Bytes())
	if qx.Sign() == 0 && qy.Sign() == 0 {
		return nil, fmt.Errorf("recovered the point at infinity")
	}

	return &ecdsa.PublicKey{Curve: curve, X: qx, Y: qy}, nil
}

// rfc6979Nonce computes a deterministic nonce per RFC 6979 (HMAC-SHA256) for
// the given private key, message hash and curve order. extra increments the
// nonce stream so callers can request successive candidates if a previous one
// produced an invalid signature.
func rfc6979Nonce(d *big.Int, hash []byte, n *big.Int, extra int) *big.Int {
	holen := sha256.Size
	rolen := (n.BitLen() + 7) / 8

	bx := append(int2octets(d, rolen), bits2octets(hash, n, rolen)...)

	v := make([]byte, holen)
	k := make([]byte, holen)
	for i := range v {
		v[i] = 0x01
	}

	mac := func(key, data []byte) []byte {
		h := hmac.New(sha256.New, key)
		h.Write(data)
		return h.Sum(nil)
	}

	k = mac(k, appendBytes(v, []byte{0x00}, bx))
	v = mac(k, v)
	k = mac(k, appendBytes(v, []byte{0x01}, bx))
	v = mac(k, v)

	skip := extra
	for {
		var t []byte
		for len(t) < rolen {
			v = mac(k, v)
			t = append(t, v...)
		}
		secret := bits2int(t, n.BitLen())
		if secret.Sign() > 0 && secret.Cmp(n) < 0 {
			if skip == 0 {
				return secret
			}
			skip--
		}
		k = mac(k, append(append([]byte{}, v...), 0x00))
		v = mac(k, v)
	}
}

func appendBytes(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// int2octets encodes x as a fixed-length big-endian byte slice.
func int2octets(x *big.Int, rolen int) []byte {
	return paddedBytes(x, rolen)
}

// bits2int converts a byte string to an integer, truncating to qlen bits.
func bits2int(in []byte, qlen int) *big.Int {
	v := new(big.Int).SetBytes(in)
	if blen := len(in) * 8; blen > qlen {
		v.Rsh(v, uint(blen-qlen))
	}
	return v
}

// bits2octets converts a hash to an octet string reduced modulo n.
func bits2octets(in []byte, n *big.Int, rolen int) []byte {
	z1 := bits2int(in, n.BitLen())
	z2 := new(big.Int).Sub(z1, n)
	if z2.Sign() < 0 {
		return int2octets(z1, rolen)
	}
	return int2octets(z2, rolen)
}
