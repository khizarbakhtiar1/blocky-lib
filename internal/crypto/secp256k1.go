package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// secp256k1Curve implements the elliptic.Curve interface for the secp256k1
// curve used by Ethereum and Bitcoin (y^2 = x^3 + 7 over the prime field).
//
// Go's generic elliptic.CurveParams arithmetic hardcodes the curve coefficient
// a = -3 (true for the NIST P-curves) and therefore cannot be used for
// secp256k1 (a = 0). This type implements the group law explicitly using
// Jacobian coordinates so that key generation, ECDSA, and public-key recovery
// produce values that are valid on the Ethereum network.
//
// NOTE: the implementation favours clarity and correctness over constant-time
// execution. It uses math/big and is therefore not hardened against timing
// side-channels. It is intended for building and signing transactions in
// trusted environments, not for use as a hardware-grade signing oracle.
type secp256k1Curve struct {
	params *elliptic.CurveParams
}

var (
	s256     *secp256k1Curve
	s256Once sync.Once
)

// S256 returns the secp256k1 curve.
func S256() elliptic.Curve {
	s256Once.Do(func() {
		p, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f", 16)
		n, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
		gx, _ := new(big.Int).SetString("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798", 16)
		gy, _ := new(big.Int).SetString("483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8", 16)

		s256 = &secp256k1Curve{
			params: &elliptic.CurveParams{
				P:       p,
				N:       n,
				B:       big.NewInt(7),
				Gx:      gx,
				Gy:      gy,
				BitSize: 256,
				Name:    "secp256k1",
			},
		}
	})
	return s256
}

// Params returns the curve parameters.
func (c *secp256k1Curve) Params() *elliptic.CurveParams {
	return c.params
}

// polynomial computes x^3 + 7 mod P.
func (c *secp256k1Curve) polynomial(x *big.Int) *big.Int {
	x3 := new(big.Int).Mul(x, x)
	x3.Mul(x3, x)
	x3.Add(x3, c.params.B)
	x3.Mod(x3, c.params.P)
	return x3
}

// IsOnCurve reports whether (x, y) lies on the curve.
func (c *secp256k1Curve) IsOnCurve(x, y *big.Int) bool {
	if x.Sign() < 0 || x.Cmp(c.params.P) >= 0 || y.Sign() < 0 || y.Cmp(c.params.P) >= 0 {
		return false
	}
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, c.params.P)
	return y2.Cmp(c.polynomial(x)) == 0
}

// affineFromJacobian converts a Jacobian point back to affine coordinates.
func (c *secp256k1Curve) affineFromJacobian(x, y, z *big.Int) (*big.Int, *big.Int) {
	if z.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}
	p := c.params.P
	zinv := new(big.Int).ModInverse(z, p)
	zinv2 := new(big.Int).Mul(zinv, zinv)
	zinv3 := new(big.Int).Mul(zinv2, zinv)

	xOut := new(big.Int).Mul(x, zinv2)
	xOut.Mod(xOut, p)
	yOut := new(big.Int).Mul(y, zinv3)
	yOut.Mod(yOut, p)
	return xOut, yOut
}

// doubleJacobian doubles a point in Jacobian coordinates for a = 0 curves.
func (c *secp256k1Curve) doubleJacobian(x, y, z *big.Int) (*big.Int, *big.Int, *big.Int) {
	p := c.params.P
	if z.Sign() == 0 || y.Sign() == 0 {
		return new(big.Int), big.NewInt(1), new(big.Int)
	}

	a := new(big.Int).Mul(x, x) // A = X1^2
	a.Mod(a, p)
	b := new(big.Int).Mul(y, y) // B = Y1^2
	b.Mod(b, p)
	cc := new(big.Int).Mul(b, b) // C = B^2
	cc.Mod(cc, p)

	// D = 2*((X1+B)^2 - A - C)
	d := new(big.Int).Add(x, b)
	d.Mul(d, d)
	d.Sub(d, a)
	d.Sub(d, cc)
	d.Lsh(d, 1)
	d.Mod(d, p)

	e := new(big.Int).Mul(big.NewInt(3), a) // E = 3*A
	e.Mod(e, p)
	f := new(big.Int).Mul(e, e) // F = E^2
	f.Mod(f, p)

	x3 := new(big.Int).Sub(f, new(big.Int).Lsh(d, 1)) // X3 = F - 2*D
	x3.Mod(x3, p)

	y3 := new(big.Int).Sub(d, x3) // Y3 = E*(D - X3) - 8*C
	y3.Mul(y3, e)
	eightC := new(big.Int).Lsh(cc, 3)
	y3.Sub(y3, eightC)
	y3.Mod(y3, p)

	z3 := new(big.Int).Mul(y, z) // Z3 = 2*Y1*Z1
	z3.Lsh(z3, 1)
	z3.Mod(z3, p)

	return x3, y3, z3
}

// addJacobian adds two Jacobian points (a = 0 curve).
func (c *secp256k1Curve) addJacobian(x1, y1, z1, x2, y2, z2 *big.Int) (*big.Int, *big.Int, *big.Int) {
	p := c.params.P
	if z1.Sign() == 0 {
		return new(big.Int).Set(x2), new(big.Int).Set(y2), new(big.Int).Set(z2)
	}
	if z2.Sign() == 0 {
		return new(big.Int).Set(x1), new(big.Int).Set(y1), new(big.Int).Set(z1)
	}

	z1z1 := new(big.Int).Mul(z1, z1)
	z1z1.Mod(z1z1, p)
	z2z2 := new(big.Int).Mul(z2, z2)
	z2z2.Mod(z2z2, p)

	u1 := new(big.Int).Mul(x1, z2z2)
	u1.Mod(u1, p)
	u2 := new(big.Int).Mul(x2, z1z1)
	u2.Mod(u2, p)

	s1 := new(big.Int).Mul(y1, z2)
	s1.Mul(s1, z2z2)
	s1.Mod(s1, p)
	s2 := new(big.Int).Mul(y2, z1)
	s2.Mul(s2, z1z1)
	s2.Mod(s2, p)

	h := new(big.Int).Sub(u2, u1)
	h.Mod(h, p)
	r := new(big.Int).Sub(s2, s1)
	r.Mod(r, p)

	if h.Sign() == 0 {
		if r.Sign() == 0 {
			// P == Q, fall back to doubling.
			return c.doubleJacobian(x1, y1, z1)
		}
		// P == -Q, result is the point at infinity.
		return new(big.Int), big.NewInt(1), new(big.Int)
	}

	i := new(big.Int).Lsh(h, 1) // I = (2*H)^2
	i.Mul(i, i)
	i.Mod(i, p)
	j := new(big.Int).Mul(h, i) // J = H*I
	j.Mod(j, p)

	rr := new(big.Int).Lsh(r, 1) // r = 2*(S2-S1)
	rr.Mod(rr, p)

	v := new(big.Int).Mul(u1, i) // V = U1*I
	v.Mod(v, p)

	x3 := new(big.Int).Mul(rr, rr) // X3 = r^2 - J - 2*V
	x3.Sub(x3, j)
	x3.Sub(x3, new(big.Int).Lsh(v, 1))
	x3.Mod(x3, p)

	y3 := new(big.Int).Sub(v, x3) // Y3 = r*(V - X3) - 2*S1*J
	y3.Mul(y3, rr)
	s1j := new(big.Int).Mul(s1, j)
	s1j.Lsh(s1j, 1)
	y3.Sub(y3, s1j)
	y3.Mod(y3, p)

	z3 := new(big.Int).Add(z1, z2) // Z3 = ((Z1+Z2)^2 - Z1Z1 - Z2Z2)*H
	z3.Mul(z3, z3)
	z3.Sub(z3, z1z1)
	z3.Sub(z3, z2z2)
	z3.Mul(z3, h)
	z3.Mod(z3, p)

	return x3, y3, z3
}

// Add returns the sum of two affine points.
func (c *secp256k1Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	var z1, z2 *big.Int
	if x1.Sign() == 0 && y1.Sign() == 0 {
		z1 = new(big.Int)
	} else {
		z1 = big.NewInt(1)
	}
	if x2.Sign() == 0 && y2.Sign() == 0 {
		z2 = new(big.Int)
	} else {
		z2 = big.NewInt(1)
	}
	x3, y3, z3 := c.addJacobian(x1, y1, z1, x2, y2, z2)
	return c.affineFromJacobian(x3, y3, z3)
}

// Double returns 2*(x1, y1).
func (c *secp256k1Curve) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	var z1 *big.Int
	if x1.Sign() == 0 && y1.Sign() == 0 {
		z1 = new(big.Int)
	} else {
		z1 = big.NewInt(1)
	}
	x3, y3, z3 := c.doubleJacobian(x1, y1, z1)
	return c.affineFromJacobian(x3, y3, z3)
}

// ScalarMult returns k*(Bx, By) where k is a big-endian integer.
func (c *secp256k1Curve) ScalarMult(bx, by *big.Int, k []byte) (*big.Int, *big.Int) {
	x, y, z := new(big.Int), big.NewInt(1), new(big.Int) // start at infinity
	for _, b := range k {
		for bit := 0; bit < 8; bit++ {
			x, y, z = c.doubleJacobian(x, y, z)
			if b&0x80 == 0x80 {
				x, y, z = c.addJacobian(x, y, z, bx, by, big.NewInt(1))
			}
			b <<= 1
		}
	}
	return c.affineFromJacobian(x, y, z)
}

// ScalarBaseMult returns k*G.
func (c *secp256k1Curve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return c.ScalarMult(c.params.Gx, c.params.Gy, k)
}

// FromECDSA exports a private key into a 32-byte big-endian binary dump.
func FromECDSA(priv *ecdsa.PrivateKey) []byte {
	if priv == nil || priv.D == nil {
		return nil
	}
	return paddedBytes(priv.D, 32)
}

// ToECDSA creates a private key with the given D value on the secp256k1 curve.
func ToECDSA(d []byte) (*ecdsa.PrivateKey, error) {
	if len(d) == 0 {
		return nil, fmt.Errorf("private key bytes must not be empty")
	}

	curve := S256()
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = new(big.Int).SetBytes(d)

	// Validate that the key is within the curve order.
	if priv.D.Cmp(curve.Params().N) >= 0 || priv.D.Sign() == 0 {
		return nil, fmt.Errorf("invalid private key: must be in range [1, N-1]")
	}

	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(priv.D.Bytes())
	return priv, nil
}

// HexToECDSA parses a private key from a hex-encoded string.
// Accepts optional "0x" prefix.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	hexkey = strings.TrimPrefix(hexkey, "0x")
	if len(hexkey) == 0 {
		return nil, fmt.Errorf("empty hex key")
	}

	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	return ToECDSA(b)
}

// paddedBytes returns the big-endian encoding of n left-padded to size bytes.
func paddedBytes(n *big.Int, size int) []byte {
	b := n.Bytes()
	if len(b) >= size {
		return b
	}
	out := make([]byte, size)
	copy(out[size-len(b):], b)
	return out
}
