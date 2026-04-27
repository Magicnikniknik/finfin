package pricing

import (
	"fmt"
	"math/big"
	"strings"
)

func ParseDecimalString(raw string) (*big.Rat, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(strings.TrimSpace(raw)); !ok {
		return nil, fmt.Errorf("%w: invalid decimal %q", ErrInvalidQuoteInput, raw)
	}
	return r, nil
}

func ApplyBps(amount *big.Rat, bps int) *big.Rat {
	multiplier := new(big.Rat).SetFrac64(int64(10000+bps), 10000)
	return new(big.Rat).Mul(amount, multiplier)
}

func ApplyFixedFee(amount, fee *big.Rat, subtract bool) *big.Rat {
	if subtract {
		return new(big.Rat).Sub(amount, fee)
	}
	return new(big.Rat).Add(amount, fee)
}

func RoundAmount(raw string, precision int, mode RoundingMode) (string, error) {
	if precision < 0 || precision > 18 {
		return "", fmt.Errorf("%w: rounding precision out of range", ErrInvalidQuoteInput)
	}
	v, err := ParseDecimalString(raw)
	if err != nil {
		return "", err
	}
	scaleInt := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision)), nil)
	scaled := new(big.Rat).Mul(v, new(big.Rat).SetInt(scaleInt))
	q := roundScaled(scaled, mode)
	final := new(big.Rat).SetFrac(q, scaleInt)
	return ratToString(final, precision), nil
}

func roundScaled(scaled *big.Rat, mode RoundingMode) *big.Int {
	n := scaled.Num()
	d := scaled.Denom()
	q := new(big.Int)
	r := new(big.Int)
	q.QuoRem(n, d, r)

	isNegative := scaled.Sign() < 0
	hasRem := r.Sign() != 0
	absR := new(big.Int).Abs(r)
	doubleRem := new(big.Int).Mul(absR, big.NewInt(2))

	switch mode {
	case RoundingFloor:
		if isNegative && hasRem {
			q.Sub(q, big.NewInt(1))
		}
	case RoundingCeil:
		if !isNegative && hasRem {
			q.Add(q, big.NewInt(1))
		}
	default: // half_up
		if doubleRem.Cmp(d) >= 0 {
			if isNegative {
				q.Sub(q, big.NewInt(1))
			} else {
				q.Add(q, big.NewInt(1))
			}
		}
	}
	return q
}

func ratToString(v *big.Rat, precision int) string {
	s := new(big.Float).SetRat(v).Text('f', precision)
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-" {
		return "0"
	}
	return s
}

func compareDecimalStrings(left, right string) (int, error) {
	l, err := ParseDecimalString(left)
	if err != nil {
		return 0, err
	}
	r, err := ParseDecimalString(right)
	if err != nil {
		return 0, err
	}
	return l.Cmp(r), nil
}
