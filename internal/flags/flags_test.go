package flags

import (
	"errors"
	"testing"

	"github.com/yupsh/repl/internal/constants"
)

// boolMarker is a sentinel boolean-flag option used by the flag tests.
type boolMarker struct{}

// probeFlags exercises a boolean flag, a value flag, and the "-NUM" shorthand.
var probeFlags = Set{
	Bool("b", "bool", boolMarker{}),
	Value("n", "num", IntMaker(func(n int) any { return n })),
	Num(IntMaker(func(n int) any { return n })),
}

func TestParsePositionalsAndFlags(t *testing.T) {
	opts, positional, err := Parse(probeFlags, Argv{"-b", "-n", "5", "file"})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 2 || opts[1] != 5 {
		t.Errorf("opts = %v", opts)
	}
	if len(positional) != 1 || positional[0] != "file" {
		t.Errorf("positional = %v", positional)
	}
}

func TestParseForms(t *testing.T) {
	cases := []struct {
		args     Argv
		wantOpts int
		wantPos  int
	}{
		{Argv{"--bool", "--num=7"}, 2, 0},
		{Argv{"--num", "8"}, 1, 0}, // long value flag takes the following argument
		{Argv{"-bn", "9"}, 2, 0},
		{Argv{"-n3"}, 1, 0},
		{Argv{"-5"}, 1, 0},     // "-NUM" shorthand is declared, so it is an option
		{Argv{"-", "x"}, 0, 2}, // bare dash is positional
	}
	for _, c := range cases {
		opts, positional, err := Parse(probeFlags, c.args)
		if err != nil {
			t.Fatalf("Parse(%v): %v", c.args, err)
		}
		if len(opts) != c.wantOpts || len(positional) != c.wantPos {
			t.Errorf("Parse(%v) = %d opts, %d pos", c.args, len(opts), len(positional))
		}
	}
}

// noNumFlags omits the "-NUM" shorthand, so "-5" is a negative-number positional.
var noNumFlags = Set{Bool("b", "bool", boolMarker{})}

func TestParseNegativePositional(t *testing.T) {
	opts, positional, err := Parse(noNumFlags, Argv{"-5"})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 0 || len(positional) != 1 || positional[0] != "-5" {
		t.Errorf("Parse(-5) = %v opts, %v pos", opts, positional)
	}
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		wantErr constants.Error
		args    Argv
	}{
		{args: Argv{"--num"}, wantErr: constants.ErrFlagNeedsValue},
		{args: Argv{"-n"}, wantErr: constants.ErrFlagNeedsValue},
		{args: Argv{"--bogus"}, wantErr: constants.ErrUnknownFlag},
		{args: Argv{"-z"}, wantErr: constants.ErrUnknownFlag},
		{args: Argv{"--num=x"}, wantErr: constants.ErrInvalidNumber},
		{args: Argv{"-n", "x"}, wantErr: constants.ErrInvalidNumber},
	}
	for _, c := range cases {
		_, _, err := Parse(probeFlags, c.args)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("Parse(%v) err = %v, want %v", c.args, err, c.wantErr)
		}
	}
}

func TestArgvConversions(t *testing.T) {
	a := Argv{"x", "y"}
	if ss := a.Strings(); ss[0] != "x" || ss[1] != "y" {
		t.Errorf("Strings = %v", ss)
	}
	if as := a.Anys(); as[0].(string) != "x" {
		t.Errorf("Anys = %v", as)
	}
}

func TestFirstOr(t *testing.T) {
	if got := FirstOr(Argv{"path"}, "."); got != "path" {
		t.Errorf("FirstOr(path) = %q", got)
	}
	if got := FirstOr(nil, "."); got != "." {
		t.Errorf("FirstOr(nil) = %q", got)
	}
}

func TestNumArg(t *testing.T) {
	if v, err := NumArg("7"); err != nil || v != 7 {
		t.Errorf("NumArg(7) = %v, %v", v, err)
	}
	if v, err := NumArg("1.5"); err != nil || v != 1.5 {
		t.Errorf("NumArg(1.5) = %v, %v", v, err)
	}
	if _, err := NumArg("x"); !errors.Is(err, constants.ErrInvalidNumber) {
		t.Errorf("NumArg(x) err = %v", err)
	}
}

func TestIntList(t *testing.T) {
	xs, err := IntList("1, 2 ,3")
	if err != nil || len(xs) != 3 || xs[2] != 3 {
		t.Errorf("IntList = %v, %v", xs, err)
	}
	if _, err := IntList("1,x"); !errors.Is(err, constants.ErrInvalidNumber) {
		t.Errorf("IntList(1,x) err = %v", err)
	}
}

func TestRangeArg(t *testing.T) {
	lo, hi, err := RangeArg("2-9")
	if err != nil || lo != 2 || hi != 9 {
		t.Errorf("RangeArg = %d,%d,%v", lo, hi, err)
	}
	if _, _, err := RangeArg("nodash"); !errors.Is(err, constants.ErrInvalidFlagValue) {
		t.Errorf("RangeArg(nodash) err = %v", err)
	}
	if _, _, err := RangeArg("x-3"); !errors.Is(err, constants.ErrInvalidNumber) {
		t.Errorf("RangeArg(x-3) err = %v", err)
	}
	if _, _, err := RangeArg("1-x"); !errors.Is(err, constants.ErrInvalidNumber) {
		t.Errorf("RangeArg(1-x) err = %v", err)
	}
}

func TestInt64Maker(t *testing.T) {
	maker := Int64Maker(func(n int64) any { return n })
	if v, err := maker("42"); err != nil || v.(int64) != 42 {
		t.Errorf("Int64Maker(42) = %v, %v", v, err)
	}
	if _, err := maker("x"); !errors.Is(err, constants.ErrInvalidNumber) {
		t.Errorf("Int64Maker(x) err = %v", err)
	}
}

func TestStrMaker(t *testing.T) {
	maker := StrMaker(func(s string) any { return s + "!" })
	if v, err := maker("hi"); err != nil || v.(string) != "hi!" {
		t.Errorf("StrMaker(hi) = %v, %v", v, err)
	}
}
