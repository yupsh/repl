package repl

import "strings"

// flagMaker converts a flag's string value into the typed option value that the
// underlying cmd-* constructor expects (e.g. "10" into head.HeadLines(10)).
type flagMaker func(value Argument) (any, error)

// flagSpec declares one command flag: its short letter (without dash), its long
// name (without dashes), and either a constant option (boolean flags) or a
// maker (value-taking flags). A numeric spec has no short or long name; it
// captures a bare "-NUM" token (the GNU head/tail shorthand) through its maker.
type flagSpec struct {
	short    string
	long     string
	takesArg bool
	numeric  bool
	opt      any
	make     flagMaker
}

// flagSet is a command's complete flag table.
type flagSet []flagSpec

// boolFlag declares a boolean flag mapping to a constant option value.
func boolFlag(short, long string, opt any) flagSpec {
	return flagSpec{short: short, long: long, opt: opt}
}

// valueFlag declares a value-taking flag mapping its value through make.
func valueFlag(short, long string, make flagMaker) flagSpec {
	return flagSpec{short: short, long: long, takesArg: true, make: make}
}

// numFlag declares the GNU "-NUM" shorthand, mapping the digits through make.
func numFlag(make flagMaker) flagSpec {
	return flagSpec{numeric: true, takesArg: true, make: make}
}

// lookupShort finds the spec for a short flag letter.
func (s flagSet) lookupShort(letter byte) (flagSpec, bool) {
	for _, spec := range s {
		if spec.short != "" && spec.short[0] == letter {
			return spec, true
		}
	}
	return flagSpec{}, false
}

// lookupLong finds the spec for a long flag name.
func (s flagSet) lookupLong(name string) (flagSpec, bool) {
	for _, spec := range s {
		if spec.long != "" && spec.long == name {
			return spec, true
		}
	}
	return flagSpec{}, false
}

// hasShort reports whether a short flag letter is declared.
func (s flagSet) hasShort(letter byte) bool {
	_, ok := s.lookupShort(letter)
	return ok
}

// numericSpec returns the "-NUM" shorthand spec, if the set declares one.
func (s flagSet) numericSpec() (flagSpec, bool) {
	for _, spec := range s {
		if spec.numeric {
			return spec, true
		}
	}
	return flagSpec{}, false
}

// argParser walks a segment's arguments, separating flags (resolved against a
// flagSet) from positionals. Pointer receiver: it is a single-pass mutable
// cursor, the sole contract of the type.
type argParser struct {
	set        flagSet
	args       Argv
	index      int
	opts       []any
	positional Argv
}

// parseArgs splits args into typed flag options and positional arguments using
// the command's flag table.
func parseArgs(set flagSet, args Argv) (opts []any, positional Argv, err error) {
	p := &argParser{set: set, args: args}
	if err := p.run(); err != nil {
		return nil, nil, err
	}
	return p.opts, p.positional, nil
}

// run consumes every argument.
func (p *argParser) run() error {
	for p.index < len(p.args) {
		if err := p.step(); err != nil {
			return err
		}
	}
	return nil
}

// step classifies and consumes the argument at the cursor.
func (p *argParser) step() error {
	arg := p.args[p.index]
	p.index++
	switch {
	case isLongFlag(arg):
		return p.consumeLong(string(arg)[2:])
	case isShortDash(arg):
		return p.consumeDash(string(arg)[1:])
	default:
		p.positional = append(p.positional, arg)
		return nil
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

// consumeDash resolves a single-dash token. A declared short flag wins (so
// comm's "-1" stays a flag); an all-digit body is the GNU "-NUM" shorthand when
// the command declares one, otherwise a negative-number positional; anything
// else falls through to short parsing, which reports an unknown flag.
func (p *argParser) consumeDash(body string) error {
	switch {
	case p.set.hasShort(body[0]):
		return p.consumeShort(body)
	case allDigits(body):
		return p.consumeNumeric(body)
	default:
		return p.consumeShort(body)
	}
}

// consumeNumeric applies the "-NUM" shorthand, or keeps the token as a
// negative-number positional when no shorthand is declared.
func (p *argParser) consumeNumeric(body string) error {
	if spec, ok := p.set.numericSpec(); ok {
		return p.addValue(spec, Argument(body))
	}
	p.positional = append(p.positional, Argument("-"+body))
	return nil
}

// allDigits reports whether s is non-empty and entirely ASCII digits.
func allDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// consumeLong resolves a "--name" or "--name=value" token.
func (p *argParser) consumeLong(body string) error {
	name, value, hasValue := strings.Cut(body, "=")
	spec, ok := p.set.lookupLong(name)
	if !ok {
		return ErrUnknownFlag.With(nil, "--"+name)
	}
	if !spec.takesArg {
		p.addBool(spec)
		return nil
	}
	return p.addValueFromLong(spec, name, value, hasValue)
}

// addValueFromLong applies a value flag whose value is either inline (--n=5) or
// the following argument (--n 5).
func (p *argParser) addValueFromLong(spec flagSpec, name, value string, hasValue bool) error {
	if hasValue {
		return p.addValue(spec, Argument(value))
	}
	next, ok := p.nextValue()
	if !ok {
		return ErrFlagNeedsValue.With(nil, "--"+name)
	}
	return p.addValue(spec, next)
}

// consumeShort resolves a "-x" cluster: a run of boolean letters, optionally
// ending in a value flag whose value is the cluster remainder or the next
// argument (-l, -lwc, -n10, -n 10).
func (p *argParser) consumeShort(body string) error {
	for i := 0; i < len(body); i++ {
		spec, ok := p.set.lookupShort(body[i])
		if !ok {
			return ErrUnknownFlag.With(nil, "-"+string(body[i]))
		}
		if spec.takesArg {
			return p.addShortValue(spec, body[i+1:])
		}
		p.addBool(spec)
	}
	return nil
}

// addShortValue applies a short value flag, taking its value from the cluster
// remainder when present or otherwise the next argument.
func (p *argParser) addShortValue(spec flagSpec, rest string) error {
	if rest != "" {
		return p.addValue(spec, Argument(rest))
	}
	next, ok := p.nextValue()
	if !ok {
		return ErrFlagNeedsValue.With(nil, "-"+spec.short)
	}
	return p.addValue(spec, next)
}

// nextValue returns the next unconsumed argument, advancing the cursor.
func (p *argParser) nextValue() (Argument, bool) {
	if p.index >= len(p.args) {
		return "", false
	}
	value := p.args[p.index]
	p.index++
	return value, true
}

// addBool appends a boolean flag's constant option.
func (p *argParser) addBool(spec flagSpec) {
	p.opts = append(p.opts, spec.opt)
}

// addValue appends a value flag's typed option produced by its maker.
func (p *argParser) addValue(spec flagSpec, value Argument) error {
	opt, err := spec.make(value)
	if err != nil {
		return err
	}
	p.opts = append(p.opts, opt)
	return nil
}
