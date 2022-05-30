package decimal128

import (
	"fmt"
)

var (
	nanText    = []byte("NaN")
	padNaNText = []byte(" NaN")
	posNaNText = []byte("+NaN")
	negInfText = []byte("-Inf")
	padInfText = []byte(" Inf")
	posInfText = []byte("+Inf")

	digitPairs = [...][2]byte{
		{'0', '0'}, {'0', '1'}, {'0', '2'}, {'0', '3'}, {'0', '4'},
		{'0', '5'}, {'0', '6'}, {'0', '7'}, {'0', '8'}, {'0', '9'},
		{'1', '0'}, {'1', '1'}, {'1', '2'}, {'1', '3'}, {'1', '4'},
		{'1', '5'}, {'1', '6'}, {'1', '7'}, {'1', '8'}, {'1', '9'},
		{'2', '0'}, {'2', '1'}, {'2', '2'}, {'2', '3'}, {'2', '4'},
		{'2', '5'}, {'2', '6'}, {'2', '7'}, {'2', '8'}, {'2', '9'},
		{'3', '0'}, {'3', '1'}, {'3', '2'}, {'3', '3'}, {'3', '4'},
		{'3', '5'}, {'3', '6'}, {'3', '7'}, {'3', '8'}, {'3', '9'},
		{'4', '0'}, {'4', '1'}, {'4', '2'}, {'4', '3'}, {'4', '4'},
		{'4', '5'}, {'4', '6'}, {'4', '7'}, {'4', '8'}, {'4', '9'},
		{'5', '0'}, {'5', '1'}, {'5', '2'}, {'5', '3'}, {'5', '4'},
		{'5', '5'}, {'5', '6'}, {'5', '7'}, {'5', '8'}, {'5', '9'},
		{'6', '0'}, {'6', '1'}, {'6', '2'}, {'6', '3'}, {'6', '4'},
		{'6', '5'}, {'6', '6'}, {'6', '7'}, {'6', '8'}, {'6', '9'},
		{'7', '0'}, {'7', '1'}, {'7', '2'}, {'7', '3'}, {'7', '4'},
		{'7', '5'}, {'7', '6'}, {'7', '7'}, {'7', '8'}, {'7', '9'},
		{'8', '0'}, {'8', '1'}, {'8', '2'}, {'8', '3'}, {'8', '4'},
		{'8', '5'}, {'8', '6'}, {'8', '7'}, {'8', '8'}, {'8', '9'},
		{'9', '0'}, {'9', '1'}, {'9', '2'}, {'9', '3'}, {'9', '4'},
		{'9', '5'}, {'9', '6'}, {'9', '7'}, {'9', '8'}, {'9', '9'},
	}
)

func (d Decimal) Format(f fmt.State, verb rune) {
	width, hasWidth := f.Width()

	if d.isSpecial() {
		pad := 0
		padSign := false
		printSign := false
		if verb != 'v' {
			printSign = f.Flag('+')
			padSign = f.Flag(' ')

			if hasWidth {
				pad = width
			}
		}

		f.Write(d.fmtSpecial(pad, printSign, padSign, f.Flag('-'), false))
		return
	}

	digs := d.digits()
	prec, hasPrec := f.Precision()

	switch verb {
	case 'e', 'E':
		if !hasPrec {
			prec = 6
		}

		digs.round(prec + 1)
		f.Write(digs.fmtE(prec, f.Flag('#'), f.Flag('+'), f.Flag(' '), true, byte(verb)))
	case 'f', 'F':
		if !hasPrec {
			prec = 6
		}

		if digs.exp < 0 {
			digs.round(digs.ndig + digs.exp + prec)
		}

		f.Write(digs.fmtF(prec, f.Flag('#'), f.Flag('+'), f.Flag(' ')))
	case 'g', 'G':
		if !hasPrec {
			if f.Flag('#') {
				prec = 6
			} else {
				prec = digs.ndig - 1
			}
		}

		digs.round(prec)
		exp := digs.exp + prec

		if exp < -4 || exp > prec {
			e := byte('e')
			if verb == 'G' {
				e = byte('E')
			}

			f.Write(digs.fmtE(prec, f.Flag('#'), f.Flag('+'), f.Flag(' '), true, e))
		} else {
			if digs.ndig == 0 {
				prec--
			} else {
				prec -= digs.ndig
			}

			f.Write(digs.fmtF(prec, f.Flag('#'), f.Flag('+'), f.Flag(' ')))
		}
	case 'v':
		prec = digs.ndig - 1
		exp := digs.exp + prec

		if exp < -4 || exp > prec {
			f.Write(digs.fmtE(prec, false, false, false, true, 'e'))
		} else {
			prec = 0
			if digs.exp < 0 {
				prec = -digs.exp
			}

			f.Write(digs.fmtF(prec, false, false, false))
		}
	default:
		fmt.Fprintf(f, "%%!%c(decimal128.Decimal=%s)", verb, d.String())
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (d Decimal) MarshalText() ([]byte, error) {
	if d.isSpecial() {
		return d.fmtSpecial(0, false, false, false, true), nil
	}

	digs := d.digits()
	prec := digs.ndig - 1
	exp := digs.exp + prec

	if exp < -4 || exp > prec {
		return digs.fmtE(prec, false, false, false, true, 'e'), nil
	}

	prec = 0
	if digs.exp < 0 {
		prec = -digs.exp
	}

	return digs.fmtF(prec, false, false, false), nil
}

func (d Decimal) String() string {
	if d.isSpecial() {
		return string(d.fmtSpecial(0, false, false, false, false))
	}

	digs := d.digits()
	prec := digs.ndig - 1
	exp := digs.exp + prec

	if exp < -4 || exp > prec {
		return string(digs.fmtE(prec, false, false, false, true, 'e'))
	}

	prec = 0
	if digs.exp < 0 {
		prec = -digs.exp
	}

	return string(digs.fmtF(prec, false, false, false))
}

func (d Decimal) digits() *digits {
	digs := &digits{
		neg: d.isNeg(),
	}

	sig, exp := d.decompose()

	if sig != (uint128{}) {
		digs.exp = int(exp - exponentBias)

		n := 0
		for sig != (uint128{}) {
			var rem uint64
			sig, rem = sig.div100()

			if n == 0 && rem == 0 {
				digs.exp += 2
				continue
			}

			pair := digitPairs[rem]

			if n == 0 && pair[1] == '0' {
				digs.exp++
				digs.dig[n] = pair[0]
				n++
				continue
			}

			if pair[0] == '0' && sig == (uint128{}) {
				digs.dig[n] = pair[1]
				n++
			} else {
				digs.dig[n], digs.dig[n+1] = pair[1], pair[0]
				n += 2
			}
		}

		for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
			digs.dig[i], digs.dig[j] = digs.dig[j], digs.dig[i]
		}

		digs.ndig = n
	}

	return digs
}

func (d Decimal) fmtSpecial(pad int, printSign, padSign, padRight, copyBuf bool) []byte {
	var buf []byte

	if d.isNaN() {
		if printSign {
			buf = posNaNText
		} else if padSign {
			buf = padNaNText
		} else {
			buf = nanText
		}
	} else {
		if d.isNeg() {
			buf = negInfText
		} else {
			if padSign && !printSign {
				buf = padInfText
			} else {
				buf = posInfText
			}
		}
	}

	if p := pad - len(buf); p > 0 {
		tmp := make([]byte, pad)

		if padRight {
			n := copy(tmp, buf)

			for i := n; i < pad; i++ {
				tmp[i] = ' '
			}
		} else {
			for i := 0; i < p; i++ {
				tmp[i] = ' '
			}

			copy(tmp[p:], buf)
		}

		buf = tmp
		copyBuf = false
	}

	if copyBuf {
		tmp := make([]byte, len(buf))
		copy(tmp, buf)
		buf = tmp
	}

	return buf
}

type digits struct {
	neg  bool
	dig  [39]byte
	exp  int
	ndig int
}

func (d *digits) fmtE(prec int, forceDP, printSign, padSign, padExp bool, e byte) []byte {
	var buf []byte

	if d.neg {
		buf = append(buf, '-')
	} else if printSign {
		buf = append(buf, '+')
	} else if padSign {
		buf = append(buf, ' ')
	}

	if d.ndig == 0 {
		buf = append(buf, '0')
	} else {
		buf = append(buf, d.dig[0])
	}

	if prec > 0 {
		buf = append(buf, '.')

		i := 0
		if d.ndig > 1 {
			buf = append(buf, d.dig[1:d.ndig]...)
			i = d.ndig - 1
		}

		for ; i < prec; i++ {
			buf = append(buf, '0')
		}
	} else if forceDP {
		buf = append(buf, '.')
	}

	buf = append(buf, e)

	exp := d.exp
	if d.ndig > 1 {
		exp += d.ndig - 1
	}

	if exp < 0 {
		exp = -exp
		buf = append(buf, '-')
	} else {
		buf = append(buf, '+')
	}

	if exp < 10 {
		if padExp {
			buf = append(buf, '0', '0'+byte(exp))
		} else {
			buf = append(buf, '0'+byte(exp))
		}
	} else if exp < 100 {
		buf = append(buf, '0'+byte(exp/10), '0'+byte(exp%10))
	} else if exp < 1000 {
		buf = append(buf, '0'+byte(exp/100), '0'+byte(exp/10%10), '0'+byte(exp%10))
	} else {
		buf = append(buf, '0'+byte(exp/1000), '0'+byte(exp/100%10), '0'+byte(exp/10%10), '0'+byte(exp%10))
	}

	return buf
}

func (d *digits) fmtF(prec int, forceDP, printSign, padSign bool) []byte {
	var buf []byte

	if d.neg {
		buf = append(buf, '-')
	} else if printSign {
		buf = append(buf, '+')
	} else if padSign {
		buf = append(buf, ' ')
	}

	dp := 0
	if d.ndig == 0 {
		buf = append(buf, '0')
	} else {
		dp = d.ndig + d.exp

		if dp > 0 {
			if d.ndig > dp {
				buf = append(buf, d.dig[:dp]...)
			} else {
				buf = append(buf, d.dig[:d.ndig]...)

				for i := d.ndig; i < dp; i++ {
					buf = append(buf, '0')
				}
			}
		} else {
			buf = append(buf, '0')
		}
	}

	if prec > 0 {
		buf = append(buf, '.')

		for ; dp < 0; dp++ {
			prec--
			buf = append(buf, '0')
		}

		i := 0
		if d.ndig > dp {
			buf = append(buf, d.dig[dp:d.ndig]...)
			i = d.ndig - dp
		}

		for ; i < prec; i++ {
			buf = append(buf, '0')
		}
	} else if forceDP {
		buf = append(buf, '.')
	}

	return buf
}

func (d *digits) round(prec int) {
	if d.ndig <= prec {
		return
	}

	if prec < 0 {
		d.exp += d.ndig
		d.ndig = 0
		return
	}

	up := false
	if d.ndig > 1 && d.ndig == prec+1 && d.dig[prec] == '5' {
		up = (d.dig[prec-1]-'0')%2 != 0
	} else {
		up = d.dig[prec] >= '5'
	}

	if up {
		i := prec - 1
		for i >= 0 && d.dig[i] == '9' {
			i--
		}

		if i == -1 {
			d.dig[0] = '1'
			d.exp += d.ndig
			d.ndig = 1
		} else {
			d.dig[i]++
			prec = i + 1
			d.exp += d.ndig - prec
			d.ndig = prec
		}
	} else {
		i := prec - 1
		for i >= 0 && d.dig[i] == '0' {
			i--
		}

		prec = i + 1
		d.exp += d.ndig - prec
		d.ndig = prec
	}
}
