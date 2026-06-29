package pipeline

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/command"
	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/flags"
)

// seg resolves and builds a single command segment, the way the line domain
// does, so the pipeline tests can assemble realistic stages.
func seg(t *testing.T, fs afero.Fs, name string, args ...string) Stage {
	t.Helper()
	b := command.Registry()[command.Name(name)]
	argv := make(flags.Argv, len(args))
	for i, a := range args {
		argv[i] = flags.Argument(a)
	}
	var (
		s   command.Segment
		err error
	)
	if b.Raw {
		s, err = b.Build(fs, argv, nil)
	} else {
		opts, positional, perr := flags.Parse(b.Flags, argv)
		if perr != nil {
			t.Fatalf("parse %s: %v", name, perr)
		}
		s, err = b.Build(fs, positional, opts)
	}
	if err != nil {
		t.Fatalf("build %s: %v", name, err)
	}
	return Stage{Name: command.Name(name), Segment: s}
}

// run plans the stages and executes the pipeline against the given stdin.
func run(t *testing.T, fs afero.Fs, stdin string, stages ...Stage) (string, error) {
	t.Helper()
	asm, err := Plan(stages, fs, strings.NewReader(stdin))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = asm.Run(context.Background(), &out)
	return out.String(), err
}

func TestPlanFirstStageSources(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/f.txt", []byte("a\nb\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		stdin  string
		want   string
		stages []Stage
	}{
		{name: "stdin default (no stages)", stdin: "hi\n", stages: nil, want: "hi\n"},
		{name: "source command", stdin: "", stages: []Stage{seg(t, fs, "echo", "hi")}, want: "hi\n"},
		{name: "literal inputs", stdin: "", stages: []Stage{seg(t, fs, "basename", "/a/b.txt")}, want: "b.txt\n"},
		{name: "file source", stdin: "", stages: []Stage{seg(t, fs, "cat", "/f.txt")}, want: "a\nb\n"},
		{name: "default filter over stdin", stdin: "x\n", stages: []Stage{seg(t, fs, "cat")}, want: "x\n"},
		{
			name:   "two stages compose",
			stdin:  "",
			stages: []Stage{seg(t, fs, "echo", "hi"), seg(t, fs, "cat")},
			want:   "hi\n",
		},
	}
	for _, c := range cases {
		got, err := run(t, fs, c.stdin, c.stages...)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestPlanRejectsMidPipeline(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/f.txt", []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name    string
		wantErr constants.Error
		stages  []Stage
	}{
		{
			name:    "source after pipe",
			stages:  []Stage{seg(t, fs, "echo", "a"), seg(t, fs, "echo", "b")},
			wantErr: constants.ErrSourceMidPipeline,
		},
		{
			name:    "files after pipe",
			stages:  []Stage{seg(t, fs, "echo", "a"), seg(t, fs, "cat", "/f.txt")},
			wantErr: constants.ErrArgsMidPipeline,
		},
		{
			name:    "inputs after pipe",
			stages:  []Stage{seg(t, fs, "echo", "a"), seg(t, fs, "basename", "/x/y")},
			wantErr: constants.ErrArgsMidPipeline,
		},
	}
	for _, c := range cases {
		_, err := run(t, fs, "", c.stages...)
		if !errors.Is(err, c.wantErr) {
			t.Errorf("%s err = %v, want %v", c.name, err, c.wantErr)
		}
	}
}

func TestRunPropagatesRuntimeError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// A malformed sed expression fails per line, surfacing as a run error.
	if _, err := run(t, fs, "", seg(t, fs, "echo", "x"), seg(t, fs, "sed", "nonsense")); err == nil {
		t.Error("expected runtime error from malformed sed expression")
	}
}
