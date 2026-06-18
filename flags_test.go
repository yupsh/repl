package repl

import (
	"errors"
	"testing"
)

// echoMarker is a sentinel boolean-flag option used by the flag tests.
type echoMarker struct{}

var probeFlags = flagSet{
	boolFlag("b", "bool", echoMarker{}),
	valueFlag("n", "num", intMaker(func(n int) any { return n })),
}

func TestParseArgsPositionalsAndFlags(t *testing.T) {
	opts, positional, err := parseArgs(probeFlags, Argv{"-b", "-n", "5", "file"})
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

func TestParseArgsForms(t *testing.T) {
	cases := []struct {
		args     Argv
		wantOpts int
		wantPos  int
	}{
		{Argv{"--bool", "--num=7"}, 2, 0},
		{Argv{"-bn", "9"}, 2, 0},
		{Argv{"-n3"}, 1, 0},
		{Argv{"-5"}, 0, 1},     // negative number is positional (no -NUM shorthand)
		{Argv{"-", "x"}, 0, 2}, // bare dash is positional
	}
	for _, c := range cases {
		opts, positional, err := parseArgs(probeFlags, c.args)
		if err != nil {
			t.Fatalf("parseArgs(%v): %v", c.args, err)
		}
		if len(opts) != c.wantOpts || len(positional) != c.wantPos {
			t.Errorf("parseArgs(%v) = %d opts, %d pos", c.args, len(opts), len(positional))
		}
	}
}

func TestParseArgsErrors(t *testing.T) {
	cases := []struct {
		args    Argv
		wantErr Error
	}{
		{Argv{"--num"}, ErrFlagNeedsValue},
		{Argv{"-n"}, ErrFlagNeedsValue},
		{Argv{"--bogus"}, ErrUnknownFlag},
		{Argv{"-z"}, ErrUnknownFlag},
		{Argv{"--num=x"}, ErrInvalidNumber},
		{Argv{"-n", "x"}, ErrInvalidNumber},
	}
	for _, c := range cases {
		_, _, err := parseArgs(probeFlags, c.args)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("parseArgs(%v) err = %v, want %v", c.args, err, c.wantErr)
		}
	}
}
