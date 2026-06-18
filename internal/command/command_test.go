package command

import (
	"errors"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/flags"
)

// build resolves a command by name, parses its flags (unless it is a raw
// command), and builds its segment — the same resolution the line domain
// performs, reduced to what the command package needs to exercise its builders
// and flag makers.
func build(t *testing.T, fs afero.Fs, name string, args ...string) (Segment, error) {
	t.Helper()
	b, ok := Registry()[Name(name)]
	if !ok {
		t.Fatalf("unknown command %q", name)
	}
	argv := make(flags.Argv, len(args))
	for i, a := range args {
		argv[i] = flags.Argument(a)
	}
	if b.Raw {
		return b.Build(fs, argv, nil)
	}
	opts, positional, err := flags.Parse(b.Flags, argv)
	if err != nil {
		return Segment{}, err
	}
	return b.Build(fs, positional, opts)
}

// TestEveryCommandBuilds exercises every registry builder and value-flag maker,
// covering each construction closure at least once.
func TestEveryCommandBuilds(t *testing.T) {
	fs := afero.NewMemMapFs()
	lines := [][]string{
		{"echo", "hi", "there"},
		{"emit", "literal", "text"},
		{"yes"},
		{"yes", "custom", "line"},
		{"seq", "1", "5"},
		{"seq", "1", "0.5", "2"},
		{"ls", "."},
		{"ls"},
		{"ls", "a.go", "b.go"}, // multiple args → echo names (gloo.SliceSource)
		{"find", ".", "--name", "*.go", "--type", "f", "--maxdepth", "2"},
		{"base64", "-d"},
		{"base64", "--decode"},
		{"basename", "-s", ".txt"},
		{"cat", "-n", "-b"},
		{"cat", "data.txt"}, // filter with a file positional → toFiles
		{"comm", "-1", "-2", "-3"},
		{"cut", "-d", ",", "-f", "1,3"},
		{"cut", "-b", "1-3", "-c", "2", "--complement"},
		{"cut", "--fields=1,2"},
		{"diff", "-u"},
		{"dirname"},
		{"head", "-n", "5"},
		{"head", "-c", "10"},
		{"hexdump", "-C"},
		{"join", "-t", ":"},
		{"json"},
		{"nl", "-b", "a", "-s", ":", "-v", "1", "-i", "2", "-w", "4", "-n", "rn"},
		{"nl", "-b", "t"},
		{"nl", "-b", "n"},
		{"paste", "-d", ",", "-s"},
		{"perl", "-n", "-p", "-a", "code"},
		{"rev"},
		{"sed", "s/a/b/"},
		{"shuf", "-n", "2"},
		{"shuf", "--seed", "5"},
		{"shuf", "-i", "1-3"},
		{"shuf", "-e", "a", "b", "c"},
		{"sort", "-r", "-n", "-u", "-f", "-R", "-b", "-V", "-h", "-M", "-s", "-k", "1", "-t", ","},
		{"split", "-d", ","},
		{"tac", "-s", ","},
		{"tail", "-n", "2", "-c", "5"},
		{"tee"},
		{"tr", "a", "b"},
		{"tr", "-d", "a"},
		{"uniq", "-d"},
		{"uniq", "-c"},
		{"wc", "-l", "-w", "-c", "-m", "-L"},
		{"xargs", "-n", "2"},
		{"git", "status", "--oneline"},
		{"exec", "echo", "hi"},
		{"grep", "pattern", "file.txt"}, // grep with files → toFiles
	}
	for _, line := range lines {
		if _, err := build(t, fs, line[0], line[1:]...); err != nil {
			t.Errorf("build %v: %v", line, err)
		}
	}
}

func TestBuilderErrors(t *testing.T) {
	fs := afero.NewMemMapFs()
	cases := []struct {
		line    []string
		wantErr constants.Error
	}{
		{[]string{"grep"}, constants.ErrMissingArgument},
		{[]string{"sed"}, constants.ErrMissingArgument},
		{[]string{"tr"}, constants.ErrMissingArgument},
		{[]string{"seq", "x"}, constants.ErrInvalidNumber},
		{[]string{"head", "-n", "x"}, constants.ErrInvalidNumber},
		{[]string{"shuf", "--seed", "x"}, constants.ErrInvalidNumber},
		{[]string{"cut", "-f", "x"}, constants.ErrInvalidNumber},
		{[]string{"shuf", "-i", "nodash"}, constants.ErrInvalidFlagValue},
		{[]string{"shuf", "-i", "x-3"}, constants.ErrInvalidNumber},
		{[]string{"shuf", "-i", "1-x"}, constants.ErrInvalidNumber},
		{[]string{"nl", "-b", "z"}, constants.ErrInvalidFlagValue},
	}
	for _, c := range cases {
		_, err := build(t, fs, c.line[0], c.line[1:]...)
		if err == nil || !strings.Contains(err.Error(), string(c.wantErr)) {
			t.Errorf("build %v err = %v, want %q", c.line, err, c.wantErr)
		}
	}
}

func TestLsListsNames(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := fs.MkdirAll("/dir", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := afero.WriteFile(fs, "/dir/file.txt", []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		positional flags.Argv
		want       bool
	}{
		{nil, false},                        // no argument → list cwd
		{flags.Argv{"/dir"}, false},         // single directory → list its entries
		{flags.Argv{"/dir/file.txt"}, true}, // single non-directory → echo name
		{flags.Argv{"a", "b"}, true},        // multiple arguments → echo names
	}
	for _, c := range cases {
		if got := lsListsNames(fs, c.positional); got != c.want {
			t.Errorf("lsListsNames(%v) = %v, want %v", c.positional, got, c.want)
		}
	}
}

func TestToFilesToLines(t *testing.T) {
	a := flags.Argv{"x", "y"}
	if fs := toFiles(a); len(fs) != 2 || string(fs[1]) != "y" {
		t.Errorf("toFiles = %v", fs)
	}
	if ls := toLines(a); len(ls) != 2 || string(ls[0]) != "x" {
		t.Errorf("toLines = %v", ls)
	}
	if ls := toLines(nil); ls != nil {
		t.Errorf("toLines(nil) = %v, want nil", ls)
	}
}

func TestNlBody(t *testing.T) {
	for _, v := range []flags.Argument{"a", "t", "n"} {
		if _, err := nlBody(v); err != nil {
			t.Errorf("nlBody(%q): %v", v, err)
		}
	}
	if _, err := nlBody("bad"); !errors.Is(err, constants.ErrInvalidFlagValue) {
		t.Error("nlBody(bad) should fail with ErrInvalidFlagValue")
	}
}
