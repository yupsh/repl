package repl

import (
	"errors"
	"testing"
)

func TestFirstOr(t *testing.T) {
	if got := firstOr(Argv{"path"}, "."); got != "path" {
		t.Errorf("firstOr(path) = %q", got)
	}
	if got := firstOr(nil, "."); got != "." {
		t.Errorf("firstOr(nil) = %q", got)
	}
}

func TestNumArg(t *testing.T) {
	if v, err := numArg("7"); err != nil || v != 7 {
		t.Errorf("numArg(7) = %v, %v", v, err)
	}
	if v, err := numArg("1.5"); err != nil || v != 1.5 {
		t.Errorf("numArg(1.5) = %v, %v", v, err)
	}
	if _, err := numArg("x"); !errors.Is(err, ErrInvalidNumber) {
		t.Errorf("numArg(x) err = %v", err)
	}
}

func TestIntList(t *testing.T) {
	xs, err := intList("1, 2 ,3")
	if err != nil || len(xs) != 3 || xs[2] != 3 {
		t.Errorf("intList = %v, %v", xs, err)
	}
	if _, err := intList("1,x"); !errors.Is(err, ErrInvalidNumber) {
		t.Errorf("intList(1,x) err = %v", err)
	}
}

func TestRangeArg(t *testing.T) {
	lo, hi, err := rangeArg("2-9")
	if err != nil || lo != 2 || hi != 9 {
		t.Errorf("rangeArg = %d,%d,%v", lo, hi, err)
	}
	if _, _, err := rangeArg("nodash"); !errors.Is(err, ErrInvalidFlagValue) {
		t.Errorf("rangeArg(nodash) err = %v", err)
	}
}

func TestArgvConversions(t *testing.T) {
	a := Argv{"x", "y"}
	if ss := a.strings(); ss[0] != "x" || ss[1] != "y" {
		t.Errorf("strings = %v", ss)
	}
	if as := a.anys(); as[0].(string) != "x" {
		t.Errorf("anys = %v", as)
	}
	if fs := toFiles(a); string(fs[1]) != "y" {
		t.Errorf("toFiles = %v", fs)
	}
}
