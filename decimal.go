// Package decimal128 provides a 128-bit decimal floating point type.
package decimal128

import (
	"math"
	"strconv"
)

const (
	exponentBias        = 6176
	maxBiasedExponent   = 12287
	maxUnbiasedExponent = maxBiasedExponent - exponentBias
	minBiasedExponent   = 0
	minUnbiasedExponent = minBiasedExponent - exponentBias
	maxDigits           = 35
)

var (
	Zero    = FromInt32(0)
	One     = FromInt32(1)
	Two     = FromInt32(2)
	Hundred = FromInt32(100)
)

const DefaultPow = 1e8

// Decimal represents a 128-bit decimal floating point value. The zero value
// for Decimal is the number +0.0.
type Decimal struct {
	lo, hi uint64
}

// Abs returns a new Decimal set to the absolute value of d.
func Abs(d Decimal) Decimal {
	return Decimal{d.lo, d.hi & 0x7fff_ffff_ffff_ffff}
}

// Inf returns a new Decimal set to positive infinity if sign >= 0, or negative
// infinity if sign < 0.
func Inf(sign int) Decimal {
	return inf(sign < 0)
}

// NaN returns a new Decimal set to the "not-a-number" value.
func NaN() Decimal {
	return nan(payloadOpNaN, 0, 0)
}

// New returns a new Decimal with the provided significand and exponent.
func New(sig int64, exp int) Decimal {
	if sig == 0 {
		return zero(false)
	}

	neg := false
	if sig < 0 {
		neg = true
		sig *= -1
	}

	if exp < minUnbiasedExponent+19 {
		return zero(neg)
	}

	if exp > maxUnbiasedExponent+39 {
		return inf(neg)
	}

	sig128, exp16 := DefaultRoundingMode.reduce64(neg, uint64(sig), int16(exp+exponentBias))

	if exp > maxBiasedExponent {
		return inf(neg)
	}

	return compose(neg, sig128, exp16)
}

func compose(neg bool, sig uint128, exp int16) Decimal {
	var hi uint64
	if sig[1] > 0x0001_ffff_ffff_ffff {
		hi = 0x6000_0000_0000_0000 | uint64(exp)<<47 | sig[1]&0x7fff_ffff_ffff
	} else {
		hi = uint64(exp)<<49 | sig[1]
	}

	if neg {
		hi |= 0x8000_0000_0000_0000
	}

	return Decimal{sig[0], hi}
}

func inf(neg bool) Decimal {
	if neg {
		return Decimal{0, 0xf800_0000_0000_0000}
	}

	return Decimal{0, 0x7800_0000_0000_0000}
}

func nan(op, lhs, rhs Payload) Decimal {
	return Decimal{uint64(op | lhs<<8 | rhs<<16), 0x7c00_0000_0000_0000}
}

func one(neg bool) Decimal {
	if neg {
		return Decimal{1, 0xb040_0000_0000_0000}
	}

	return Decimal{1, 0x3040_0000_0000_0000}
}

func zero(neg bool) Decimal {
	if neg {
		return Decimal{0, 0x8000_0000_0000_0000}
	}

	return Decimal{}
}

// Canonical returns the result of converting d into its canonical
// representation. Many values have multiple possible ways of being represented
// as a Decimal. Canonical converts each of these into a single representation.
//
// If d is ±Inf or NaN, the canonical representation consists of only the bits
// required to represent the respective special floating point value, with all
// other bits set to 0. For NaN values this also removes any payload it may
// have had.
//
// If d is finite, the canonical representation is calculated as the
// representation with an exponent closest to zero that still accurately stores
// all non-zero digits the value has.
func (d Decimal) Canonical() Decimal {
	if d.isSpecial() {
		if d.IsNaN() {
			return nan(0, 0, 0)
		}

		return inf(d.Signbit())
	}

	sig, exp := d.decompose()

	for exp > exponentBias {
		tmp := sig.mul64(10)

		if tmp[1] > 0x0002_7fff_ffff_ffff {
			break
		}

		sig = tmp
		exp--
	}

	for exp < exponentBias {
		tmp, rem := sig.div10()

		if rem != 0 {
			break
		}

		sig = tmp
		exp++
	}

	return compose(d.Signbit(), sig, exp)
}

// IsInf reports whether d is an infinity. If sign > 0, IsInf reports whether
// d is positive infinity. If sign < 0, IsInf reports whether d is negative
// infinity. If sign == 0, IsInf reports whether d is either infinity.
func (d Decimal) IsInf(sign int) bool {
	if !d.isInf() {
		return false
	}

	if sign == 0 {
		return true
	}

	if sign > 0 {
		return !d.Signbit()
	}

	return d.Signbit()
}

// IsNaN reports whether d is a "not-a-number" value.
func (d Decimal) IsNaN() bool {
	return d.hi&0x7c00_0000_0000_0000 == 0x7c00_0000_0000_0000
}

// Neg returns d with its sign negated.
func (d Decimal) Neg() Decimal {
	return Decimal{d.lo, d.hi ^ 0x8000_0000_0000_0000}
}

// Sign returns:
//
//	-1 if d <   0
//	 0 if d is ±0
//	+1 if d >   0
//
// It panics if d is NaN.
func (d Decimal) Sign() int {
	if d.IsNaN() {
		panic("Decimal(NaN).Sign()")
	}

	if d.IsZero() {
		return 0
	}

	if d.Signbit() {
		return -1
	}

	return 1
}

// IsPositive return
//
//	true if d > 0
//	false if d == 0
//	false if d < 0
func (d Decimal) IsPositive() bool {
	return d.Sign() == 1
}

// IsNegative return
//
//	true if d < 0
//	false if d == 0
//	false if d > 0
func (d Decimal) IsNegative() bool {
	return d.Sign() == -1
}

// Signbit reports whether d is negative or negative zero.
func (d Decimal) Signbit() bool {
	return d.hi&0x8000_0000_0000_0000 == 0x8000_0000_0000_0000
}

func (d Decimal) decompose() (uint128, int16) {
	var sig uint128
	var exp int16

	if d.hi&0x6000_0000_0000_0000 == 0x6000_0000_0000_0000 {
		sig = uint128{d.lo, d.hi&0x7fff_ffff_ffff | 0x0002_0000_0000_0000}
		exp = int16(d.hi & 0x1fff_8000_0000_0000 >> 47)
	} else {
		sig = uint128{d.lo, d.hi & 0x0001_ffff_ffff_ffff}
		exp = int16(d.hi & 0x7ffe_0000_0000_0000 >> 49)
	}

	return sig, exp
}

func (d Decimal) isInf() bool {
	return d.hi&0x7c00_0000_0000_0000 == 0x7800_0000_0000_0000
}

func (d Decimal) isSpecial() bool {
	return d.hi&0x7800_0000_0000_0000 == 0x7800_0000_0000_0000
}

func (d Decimal) Percentage() string {
	if d.Equal(Zero) {
		return "0"
	}
	if d.isInf() && d.IsPositive() {
		return "inf%"
	}
	if d.isInf() && d.IsNegative() {
		return "-inf%"
	}
	return strconv.FormatFloat(d.Float64()/DefaultPow*100, 'f', -1, 64) + "%"
}

func (d Decimal) FormatPercentage(prec int) string {
	if d.Equal(Zero) {
		return "0"
	}
	if d.isInf() && d.IsPositive() {
		return "inf%"
	}
	if d.isInf() && d.IsNegative() {
		return "-inf%"
	}
	pow := math.Pow10(prec)
	result := strconv.FormatFloat(
		math.Trunc(d.Float64()/DefaultPow*pow*100.)/pow, 'f', prec, 64)
	return result + "%"
}
