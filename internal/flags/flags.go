// Package flags translates a command segment's Unix-style flag tokens into the
// typed option values that the underlying gloo-foo/cmd-* constructors expect.
//
// A command declares its flags as a Set of Spec entries (short letter, long
// name, and either a constant option for boolean flags or a Maker for
// value-taking flags). Parse walks a segment's arguments against that Set,
// separating typed options from positionals, and honours short clusters
// (-lwc), inline values (-n10, --num=5), the GNU "-NUM" shorthand, and "--"
// long forms. The package holds no knowledge of the REPL or of any specific
// command; it is the reusable flag-translation layer behind the line domain.
package flags

import (
	"strings"

	"github.com/yupsh/repl/internal/constants"
)

// Argument is a single parsed token of a command segment.
type Argument string

// Argv is the ordered argument list of one command segment.
type Argv []Argument

// Strings converts the arguments to a plain string slice.
func (a Argv) Strings() []string {
	out := make([]string, len(a))
	for i, arg := range a {
		out[i] = string(arg)
	}
	return out
}

// Anys converts the arguments to an []any (for variadic ...any constructors
// that classify their inputs, such as Exec).
func (a Argv) Anys() []any {
	out := make([]any, len(a))
	for i, arg := range a {
		out[i] = string(arg)
	}
	return out
}

// Maker converts a flag's string value into the typed option value that the
// underlying cmd-* constructor expects (e.g. "10" into head.HeadLines(10)).
type Maker func(value Argument) (any, error)

// ShortFlag is a flag's single-letter short name, without its dash (the "n" of
// "-n").
type ShortFlag string

// LongFlag is a flag's long name, without its dashes (the "num" of "--num").
type LongFlag string

// Spec declares one command flag: its short letter (without dash), its long
// name (without dashes), and either a constant option (boolean flags) or a
// maker (value-taking flags). A numeric spec has no short or long name; it
// captures a bare "-NUM" token (the GNU head/tail shorthand) through its maker.
type Spec struct {
	opt       any
	make      Maker
	short     ShortFlag
	long      LongFlag
	hasArg    bool
	isNumeric bool
}

// Set is a command's complete flag table.
type Set []Spec

// Bool declares a boolean flag mapping to a constant option value.
func Bool(short ShortFlag, long LongFlag, opt any) Spec {
	return Spec{short: short, long: long, opt: opt}
}

// Value declares a value-taking flag mapping its value through maker.
func Value(short ShortFlag, long LongFlag, maker Maker) Spec {
	return Spec{short: short, long: long, hasArg: true, make: maker}
}

// Num declares the GNU "-NUM" shorthand, mapping the digits through maker.
func Num(maker Maker) Spec {
	return Spec{isNumeric: true, hasArg: true, make: maker}
}

// lookupShort finds the spec for a short flag letter.
func (s Set) lookupShort(letter byte) (Spec, bool) {
	for _, spec := range s {
		if spec.short != "" && spec.short[0] == letter {
			return spec, true
		}
	}
	return Spec{}, false
}

// lookupLong finds the spec for a long flag name.
func (s Set) lookupLong(name LongFlag) (Spec, bool) {
	for _, spec := range s {
		if spec.long != "" && spec.long == name {
			return spec, true
		}
	}
	return Spec{}, false
}

// hasShort reports whether a short flag letter is declared.
func (s Set) hasShort(letter byte) bool {
	_, ok := s.lookupShort(letter)
	return ok
}

// numericSpec returns the "-NUM" shorthand spec, if the set declares one.
func (s Set) numericSpec() (Spec, bool) {
	for _, spec := range s {
		if spec.isNumeric {
			return spec, true
		}
	}
	return Spec{}, false
}

// parser walks a segment's arguments, separating flags (resolved against a
// Set) from positionals. It is an immutable single-pass cursor: every step
// takes the parser by value and returns the advanced parser, so no method
// mutates a shared state.
type parser struct {
	set        Set
	args       Argv
	opts       []any
	positional Argv
	index      int
}

// Parse splits args into typed flag options and positional arguments using the
// command's flag table.
func Parse(set Set, args Argv) (opts []any, positional Argv, err error) {
	p, err := parser{set: set, args: args}.run()
	if err != nil {
		return nil, nil, err
	}
	return p.opts, p.positional, nil
}

// run consumes every argument.
func (p parser) run() (parser, error) {
	for p.index < len(p.args) {
		next, err := p.step()
		if err != nil {
			return parser{}, err
		}
		p = next
	}
	return p, nil
}

// step classifies and consumes the argument at the cursor.
func (p parser) step() (parser, error) {
	arg := p.args[p.index]
	p.index++
	switch {
	case isLongFlag(arg):
		return p.consumeLong(string(arg)[2:])
	case isShortDash(arg):
		return p.consumeDash(string(arg)[1:])
	default:
		p.positional = append(p.positional, arg)
		return p, nil
	}
}

// isLongFlag reports whether arg is a "--name" token.
func isLongFlag(arg Argument) bool {
	return len(arg) > 2 && arg[0] == '-' && arg[1] == '-'
}

// isShortDash reports whether arg is a single-dash token like "-x" or "-5" (and
// not a bare "-" or a long "--name").
func isShortDash(arg Argument) bool {
	return len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'
}

// dashBody is the text of a single-dash token after its dash (the "n10" of
// "-n10", the "5" of "-5").
type dashBody string

// consumeDash resolves a single-dash token. A declared short flag wins (so
// comm's "-1" stays a flag); an all-digit body is the GNU "-NUM" shorthand when
// the command declares one, otherwise a negative-number positional; anything
// else falls through to short parsing, which reports an unknown flag.
func (p parser) consumeDash(body string) (parser, error) {
	switch {
	case p.set.hasShort(body[0]):
		return p.consumeShort(body)
	case allDigits(dashBody(body)):
		return p.consumeNumeric(body)
	default:
		return p.consumeShort(body)
	}
}

// consumeNumeric applies the "-NUM" shorthand, or keeps the token as a
// negative-number positional when no shorthand is declared.
func (p parser) consumeNumeric(body string) (parser, error) {
	if spec, ok := p.set.numericSpec(); ok {
		return p.addValue(spec, Argument(body))
	}
	p.positional = append(p.positional, Argument("-"+body))
	return p, nil
}

// allDigits reports whether a dash token's body is non-empty and entirely
// ASCII digits.
func allDigits(body dashBody) bool {
	for i := 0; i < len(body); i++ {
		if body[i] < '0' || body[i] > '9' {
			return false
		}
	}
	return body != ""
}

// consumeLong resolves a "--name" or "--name=value" token.
func (p parser) consumeLong(body string) (parser, error) {
	name, value, hasValue := strings.Cut(body, "=")
	spec, ok := p.set.lookupLong(LongFlag(name))
	if !ok {
		return parser{}, constants.ErrUnknownFlag.With(nil, "--"+name)
	}
	if !spec.hasArg {
		return p.addBool(spec), nil
	}
	return p.addValueFromLong(spec, name, value, hasValue)
}

// addValueFromLong applies a value flag whose value is either inline (--n=5) or
// the following argument (--n 5).
func (p parser) addValueFromLong(spec Spec, name, value string, hasValue bool) (parser, error) {
	if hasValue {
		return p.addValue(spec, Argument(value))
	}
	next, rest, ok := p.nextValue()
	if !ok {
		return parser{}, constants.ErrFlagNeedsValue.With(nil, "--"+name)
	}
	return rest.addValue(spec, next)
}

// consumeShort resolves a "-x" cluster: a run of boolean letters, optionally
// ending in a value flag whose value is the cluster remainder or the next
// argument (-l, -lwc, -n10, -n 10).
func (p parser) consumeShort(body string) (parser, error) {
	for i := 0; i < len(body); i++ {
		spec, ok := p.set.lookupShort(body[i])
		if !ok {
			return parser{}, constants.ErrUnknownFlag.With(nil, "-"+string(body[i]))
		}
		if spec.hasArg {
			return p.addShortValue(spec, body[i+1:])
		}
		p = p.addBool(spec)
	}
	return p, nil
}

// addShortValue applies a short value flag, taking its value from the cluster
// remainder when present or otherwise the next argument.
func (p parser) addShortValue(spec Spec, rest string) (parser, error) {
	if rest != "" {
		return p.addValue(spec, Argument(rest))
	}
	next, advanced, ok := p.nextValue()
	if !ok {
		return parser{}, constants.ErrFlagNeedsValue.With(nil, "-"+string(spec.short))
	}
	return advanced.addValue(spec, next)
}

// nextValue returns the next unconsumed argument and the advanced parser.
func (p parser) nextValue() (Argument, parser, bool) {
	if p.index >= len(p.args) {
		return "", p, false
	}
	value := p.args[p.index]
	p.index++
	return value, p, true
}

// addBool appends a boolean flag's constant option.
func (p parser) addBool(spec Spec) parser {
	p.opts = append(p.opts, spec.opt)
	return p
}

// addValue appends a value flag's typed option produced by its maker.
func (p parser) addValue(spec Spec, value Argument) (parser, error) {
	opt, err := spec.make(value)
	if err != nil {
		return parser{}, err
	}
	p.opts = append(p.opts, opt)
	return p, nil
}
