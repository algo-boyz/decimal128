package decimal128

import (
	"testing"

	"github.com/cockroachdb/apd/v3"
)

func TestExp(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   49,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Exp(decval)

		val.Big(bigval)

		bigctx.Exp(bigres, bigval)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Exp(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestExp10(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   42,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	bigln10 := new(apd.Decimal)
	bigctx.Ln(bigln10, apd.New(10, 0))

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Exp10(decval)

		val.Big(bigval)

		bigctx.Mul(bigres, bigval, bigln10)
		bigctx.Exp(bigres, bigres)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Exp10(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestExp2(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   40,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	bigln2 := new(apd.Decimal)
	bigctx.Ln(bigln2, apd.New(2, 0))

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Exp2(decval)

		val.Big(bigval)

		bigctx.Mul(bigres, bigval, bigln2)
		bigctx.Exp(bigres, bigres)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Exp2(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestLog(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   39,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Log(decval)

		val.Big(bigval)

		bigctx.Ln(bigres, bigval)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Log(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestLog10(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   39,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Log10(decval)

		val.Big(bigval)

		bigctx.Log10(bigres, bigval)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Log10(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestLog2(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   39,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	bigln2 := new(apd.Decimal)
	bigctx.Ln(bigln2, apd.New(2, 0))

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Log2(decval)

		val.Big(bigval)

		bigctx.Ln(bigres, bigval)
		bigctx.Quo(bigres, bigres, bigln2)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Log2(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func TestSqrt(t *testing.T) {
	t.Parallel()

	initDecimalValues()

	bigval := new(apd.Decimal)
	bigres := new(apd.Decimal)
	bigctx := apd.Context{
		Precision:   39,
		MaxExponent: 6145,
		MinExponent: -6176,
		Rounding:    apd.RoundHalfEven,
	}

	for _, val := range decimalValues {
		decval := val.Decimal()
		res := Sqrt(decval)

		val.Big(bigval)

		bigctx.Sqrt(bigres, bigval)

		if !decimalsEqual(res, bigres, bigctx.Rounding) {
			t.Errorf("Sqrt(%v) = %v, want %v", val, res, bigres)
		}
	}
}

func BenchmarkExp(b *testing.B) {
	initDecimalValues()

	decvals := make([]Decimal, len(decimalValues))
	for i, val := range decimalValues {
		decvals[i] = val.Decimal()
	}

	b.Run("Exp", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Exp(decval)
			}
		}
	})

	b.Run("Exp10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Exp10(decval)
			}
		}
	})

	b.Run("Exp2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Exp2(decval)
			}
		}
	})
}

func BenchmarkLog(b *testing.B) {
	initDecimalValues()

	decvals := make([]Decimal, len(decimalValues))
	for i, val := range decimalValues {
		decvals[i] = val.Decimal()
	}

	b.Run("Log", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Log(decval)
			}
		}
	})

	b.Run("Log10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Log10(decval)
			}
		}
	})

	b.Run("Log2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, decval := range decvals {
				Log2(decval)
			}
		}
	})
}
