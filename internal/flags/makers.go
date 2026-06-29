package flags

import (
	"strconv"
	"strings"

	"github.com/yupsh/repl/internal/constants"
)

// IntMaker builds a value-flag Maker for an int-typed option.
func IntMaker(maker func(int) any) Maker {
	return func(value Argument) (any, error) {
		n, err := AtoiArg(value)
		if err != nil {
			return nil, err
		}
		return maker(n), nil
	}
}

// Int64Maker builds a value-flag Maker for an int64-typed option.
func Int64Maker(maker func(int64) any) Maker {
	return func(value Argument) (any, error) {
		n, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, constants.ErrInvalidNumber.With(err, string(value))
		}
		return maker(n), nil
	}
}

// StrMaker builds a value-flag Maker for a string-typed option.
func StrMaker(maker func(string) any) Maker {
	return func(value Argument) (any, error) {
		return maker(string(value)), nil
	}
}

// AtoiArg parses an argument as a base-10 integer.
func AtoiArg(value Argument) (int, error) {
	n, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, constants.ErrInvalidNumber.With(err, string(value))
	}
	return n, nil
}

// NumArg parses an argument as an int when integral, otherwise a float64. Seq
// accepts both forms.
func NumArg(value Argument) (any, error) {
	if n, err := strconv.Atoi(string(value)); err == nil {
		return n, nil
	}
	f, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return nil, constants.ErrInvalidNumber.With(err, string(value))
	}
	return f, nil
}

// IntList parses a comma-separated integer list like "1,3,5".
func IntList(value Argument) ([]int, error) {
	parts := strings.Split(string(value), ",")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, err := AtoiArg(Argument(strings.TrimSpace(p)))
		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

// RangeArg parses a "lo-hi" integer range.
func RangeArg(value Argument) (lo, hi int, err error) {
	loStr, hiStr, ok := strings.Cut(string(value), "-")
	if !ok {
		return 0, 0, constants.ErrInvalidFlagValue.With(nil, string(value))
	}
	if lo, err = AtoiArg(Argument(loStr)); err != nil {
		return 0, 0, err
	}
	if hi, err = AtoiArg(Argument(hiStr)); err != nil {
		return 0, 0, err
	}
	return lo, hi, nil
}

// FirstOr returns the first positional as a string, or def when there are none.
func FirstOr(positional Argv, def string) string {
	if len(positional) > 0 {
		return string(positional[0])
	}
	return def
}
