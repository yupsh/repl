package command

import (
	"context"
	"errors"
	"testing"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/flags"
)

// runArgv resolves argv against the registry and runs the resulting command
// against empty input, returning its output lines.
func runArgv(t *testing.T, fs afero.Fs, argv ...string) []string {
	t.Helper()
	cmd, err := commandFromArgv(fs, Registry(), argv)
	if err != nil {
		t.Fatalf("commandFromArgv(%v): %v", argv, err)
	}
	lines, err := cmd.Execute(context.Background(), gloo.StreamOf[[]byte]()).Collect()
	if err != nil {
		t.Fatalf("execute(%v): %v", argv, err)
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = string(l)
	}
	return out
}

func equalLines(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCommandFromArgv_SegmentKinds(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/f.txt", []byte("x\ny\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		argv []string
		want []string
	}{
		{"source segment (echo)", []string{"echo", "a", "b"}, []string{"a b"}},
		{"files segment (cat)", []string{"cat", "/f.txt"}, []string{"x", "y"}},
		{"default empty source (grep, no files)", []string{"grep", "z"}, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := runArgv(t, fs, c.argv...); !equalLines(got, c.want) {
				t.Errorf("runArgv(%v) = %q, want %q", c.argv, got, c.want)
			}
		})
	}
}

func TestSegmentSource_LiteralInputs(t *testing.T) {
	// The inputs branch of segmentSource sources a segment's literal lines.
	src := segmentSource(afero.NewMemMapFs(), Segment{Inputs: [][]byte{[]byte("alpha"), []byte("beta")}})
	lines, err := src.Stream(context.Background()).Collect()
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	got := make([]string, len(lines))
	for i, l := range lines {
		got[i] = string(l)
	}
	if !equalLines(got, []string{"alpha", "beta"}) {
		t.Errorf("inputs source = %q, want [alpha beta]", got)
	}
}

func TestCommandFromArgv_Errors(t *testing.T) {
	fs := afero.NewMemMapFs()
	reg := Registry()
	cases := []struct {
		want error
		name string
		argv []string
	}{
		{constants.ErrEmptyCommand, "empty", nil},
		{constants.ErrUnknownCommand, "unknown", []string{"nope"}},
		{constants.ErrMissingArgument, "build error (grep no pattern)", []string{"grep"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := commandFromArgv(fs, reg, c.argv); !errors.Is(err, c.want) {
				t.Errorf("commandFromArgv(%v) err = %v, want %v", c.argv, err, c.want)
			}
		})
	}
}

func TestCommandFromArgv_ParseError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// head -n wants an integer; a non-numeric value fails flag parsing.
	if _, err := commandFromArgv(fs, Registry(), []string{"head", "-n", "abc"}); err == nil {
		t.Fatal("expected a flag parse error, got nil")
	}
}

func TestXargsFactory_RegistryFirstThenSubprocess(t *testing.T) {
	fs := afero.NewMemMapFs()
	factory := xargsFactory(fs)

	// Known command resolves via the registry and runs.
	known := factory([]string{"echo", "hi"})
	lines, err := known.Execute(context.Background(), gloo.StreamOf[[]byte]()).Collect()
	if err != nil {
		t.Fatalf("known.Execute: %v", err)
	}
	if len(lines) != 1 || string(lines[0]) != "hi" {
		t.Errorf("known = %q, want [hi]", lines)
	}

	// Unknown command falls back to a subprocess command (constructed, not run).
	if factory([]string{"definitely-not-a-registered-command"}) == nil {
		t.Error("expected a subprocess fallback command, got nil")
	}
}

func TestBuildXargs_RegroupAndExecBothSetCommand(t *testing.T) {
	fs := afero.NewMemMapFs()

	regroup, err := buildXargs(fs, nil, nil)
	if err != nil || regroup.Command == nil {
		t.Fatalf("regroup buildXargs: seg=%+v err=%v", regroup, err)
	}

	exec, err := buildXargs(fs, flags.Argv{"echo"}, nil)
	if err != nil || exec.Command == nil {
		t.Fatalf("exec buildXargs: seg=%+v err=%v", exec, err)
	}
}
