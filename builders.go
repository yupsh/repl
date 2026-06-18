package repl

import (
	"strconv"
	"strings"

	gloo "github.com/gloo-foo/framework"
)

// filterMaker builds a transform command from its parsed options.
type filterMaker func(opts []any) gloo.Command[[]byte, []byte]

// commandMaker builds a transform command from positionals and options, used by
// commands that consume their own positional arguments (exec, git, perl).
type commandMaker func(positional Argv, opts []any) gloo.Command[[]byte, []byte]

// filter adapts a filterMaker into a buildFunc, routing positional arguments to
// the pipeline source as files.
func filter(make filterMaker) buildFunc {
	return func(_ environment, positional Argv, opts []any) (segment, error) {
		return segment{command: make(opts), files: toFiles(positional)}, nil
	}
}

// command adapts a commandMaker into a buildFunc; the command reads the pipeline
// input directly, so positionals are not routed to a file source.
func command(make commandMaker) buildFunc {
	return func(_ environment, positional Argv, opts []any) (segment, error) {
		return segment{command: make(positional, opts)}, nil
	}
}

// literal adapts a filterMaker into a buildFunc whose positional arguments are
// the command's input lines rather than filenames — used by path processors
// (basename, dirname) that transform their arguments as data.
func literal(make filterMaker) buildFunc {
	return func(_ environment, positional Argv, opts []any) (segment, error) {
		return segment{command: make(opts), inputs: toLines(positional)}, nil
	}
}

// intMaker builds a value-flag maker for an int-typed option.
func intMaker(make func(int) any) flagMaker {
	return func(value Argument) (any, error) {
		n, err := atoiArg(value)
		if err != nil {
			return nil, err
		}
		return make(n), nil
	}
}

// int64Maker builds a value-flag maker for an int64-typed option.
func int64Maker(make func(int64) any) flagMaker {
	return func(value Argument) (any, error) {
		n, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return nil, ErrInvalidNumber.With(err, string(value))
		}
		return make(n), nil
	}
}

// strMaker builds a value-flag maker for a string-typed option.
func strMaker(make func(string) any) flagMaker {
	return func(value Argument) (any, error) {
		return make(string(value)), nil
	}
}

// atoiArg parses an argument as a base-10 integer.
func atoiArg(value Argument) (int, error) {
	n, err := strconv.Atoi(string(value))
	if err != nil {
		return 0, ErrInvalidNumber.With(err, string(value))
	}
	return n, nil
}

// numArg parses an argument as an int when integral, otherwise a float64. Seq
// accepts both forms.
func numArg(value Argument) (any, error) {
	if n, err := strconv.Atoi(string(value)); err == nil {
		return n, nil
	}
	f, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return nil, ErrInvalidNumber.With(err, string(value))
	}
	return f, nil
}

// intList parses a comma-separated integer list like "1,3,5".
func intList(value Argument) ([]int, error) {
	parts := strings.Split(string(value), ",")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, err := atoiArg(Argument(strings.TrimSpace(p)))
		if err != nil {
			return nil, err
		}
		out[i] = n
	}
	return out, nil
}

// rangeArg parses a "lo-hi" integer range.
func rangeArg(value Argument) (lo, hi int, err error) {
	loStr, hiStr, ok := strings.Cut(string(value), "-")
	if !ok {
		return 0, 0, ErrInvalidFlagValue.With(nil, string(value))
	}
	if lo, err = atoiArg(Argument(loStr)); err != nil {
		return 0, 0, err
	}
	if hi, err = atoiArg(Argument(hiStr)); err != nil {
		return 0, 0, err
	}
	return lo, hi, nil
}

// firstOr returns the first positional as a string, or def when there are none.
func firstOr(positional Argv, def string) string {
	if len(positional) > 0 {
		return string(positional[0])
	}
	return def
}
