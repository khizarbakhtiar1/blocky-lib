package types

import (
	"fmt"
	"math/big"
)

// Unit represents different denominations for converting values
type Unit int

const (
	Wei Unit = iota
	Kwei
	Mwei
	Gwei
	Szabo
	Finney
	Ether
)

var unitMultipliers = map[Unit]*big.Int{
	Wei:    big.NewInt(1),
	Kwei:   big.NewInt(1e3),
	Mwei:   big.NewInt(1e6),
	Gwei:   big.NewInt(1e9),
	Szabo:  big.NewInt(1e12),
	Finney: big.NewInt(1e15),
	Ether:  big.NewInt(1e18),
}

// ToWei converts a value in the specified unit to Wei
func ToWei(value *big.Int, unit Unit) *big.Int {
	multiplier, ok := unitMultipliers[unit]
	if !ok {
		panic(fmt.Sprintf("unknown unit: %d", unit))
	}
	result := new(big.Int).Mul(value, multiplier)
	return result
}

// FromWei converts a value in Wei to the specified unit
func FromWei(value *big.Int, unit Unit) *big.Int {
	multiplier, ok := unitMultipliers[unit]
	if !ok {
		panic(fmt.Sprintf("unknown unit: %d", unit))
	}
	result := new(big.Int).Div(value, multiplier)
	return result
}

// EtherToWei converts ether to wei
func EtherToWei(ether *big.Int) *big.Int {
	return ToWei(ether, Ether)
}

// WeiToEther converts wei to ether
func WeiToEther(wei *big.Int) *big.Int {
	return FromWei(wei, Ether)
}

// GweiToWei converts gwei to wei
func GweiToWei(gwei *big.Int) *big.Int {
	return ToWei(gwei, Gwei)
}

// WeiToGwei converts wei to gwei
func WeiToGwei(wei *big.Int) *big.Int {
	return FromWei(wei, Gwei)
}

// ParseEther creates a big.Int from a string in ether (e.g., "1.5")
func ParseEther(s string) (*big.Int, error) {
	// Parse as float and convert to wei
	f, ok := new(big.Float).SetString(s)
	if !ok {
		return nil, fmt.Errorf("invalid ether value: %s", s)
	}

	// Multiply by 1e18
	multiplier := new(big.Float).SetInt(big.NewInt(1e18))
	f.Mul(f, multiplier)

	// Convert to big.Int
	result := new(big.Int)
	f.Int(result)

	return result, nil
}

// FormatEther formats wei as ether with decimal places
func FormatEther(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	ether := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetInt(big.NewInt(1e18))
	ether.Quo(ether, divisor)

	return ether.Text('f', 18)
}

// FormatGwei formats wei as gwei
func FormatGwei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	gwei := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetInt(big.NewInt(1e9))
	gwei.Quo(gwei, divisor)

	return gwei.Text('f', 9)
}
