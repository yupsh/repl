package repl

import (
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// buildStage0 parses a single line and builds its first stage.
func buildStage0(t *testing.T, line string) error {
	t.Helper()
	eng := New(strings.NewReader(""), &strings.Builder{}, &strings.Builder{}, afero.NewMemMapFs(), testHome)
	segs, err := parseLine(Line(line))
	if err != nil {
		return err
	}
	_, err = eng.buildStage(segs[0])
	return err
}

// TestEveryCommandBuilds exercises every registry builder and value-flag maker,
// covering each construction closure at least once.
func TestEveryCommandBuilds(t *testing.T) {
	lines := []string{
		"echo hi there",
		"emit literal text",
		"yes",
		"yes custom line",
		"seq 1 5",
		"seq 1 0.5 2",
		"ls .",
		"ls",
		"find . --name *.go --type f --maxdepth 2",
		"base64 -d",
		"base64 --decode",
		"basename -s .txt",
		"cat -n -b",
		"comm -1 -2 -3",
		"cut -d , -f 1,3",
		"cut -b 1-3 -c 2 --complement",
		"cut --fields=1,2",
		"diff -u",
		"dirname",
		"head -n 5",
		"head -c 10",
		"hexdump -C",
		"join -t :",
		"json",
		"nl -b a -s : -v 1 -i 2 -w 4 -n rn",
		"nl -b t",
		"nl -b n",
		"paste -d , -s",
		"perl -n -p -a code",
		"rev",
		"sed s/a/b/",
		"sed s|a|b|",
		"shuf -n 2",
		"shuf --seed 5",
		"shuf -i 1-3",
		"shuf -e a b c",
		"sort -r -n -u -f -R -b -V -h -M -s -k 1 -t ,",
		"split -d ,",
		"tac -s ,",
		"tail -n 2 -c 5",
		"tee",
		"tr a b",
		"tr -d a",
		"uniq -d",
		"uniq -c",
		"wc -l -w -c -m -L",
		"xargs -n 2",
		"git status --oneline",
		"exec echo hi",
	}
	for _, line := range lines {
		if err := buildStage0(t, line); err != nil {
			t.Errorf("build %q: %v", line, err)
		}
	}
}

func TestBuilderErrors(t *testing.T) {
	cases := []struct {
		line    string
		wantErr Error
	}{
		{"grep", ErrMissingArgument},
		{"sed", ErrMissingArgument},
		{"tr", ErrMissingArgument},
		{"seq x", ErrInvalidNumber},
		{"head -n x", ErrInvalidNumber},
		{"shuf --seed x", ErrInvalidNumber},
		{"cut -f x", ErrInvalidNumber},
		{"shuf -i nodash", ErrInvalidFlagValue},
		{"shuf -i x-3", ErrInvalidNumber},
		{"shuf -i 1-x", ErrInvalidNumber},
		{"nl -b z", ErrInvalidFlagValue},
	}
	for _, c := range cases {
		err := buildStage0(t, c.line)
		if err == nil || !strings.Contains(err.Error(), string(c.wantErr)) {
			t.Errorf("build %q err = %v, want %q", c.line, err, c.wantErr)
		}
	}
}

func TestTrSet2AndDelete(t *testing.T) {
	fs := afero.NewMemMapFs()
	if got := mustExec(t, fs, "echo abcabc | tr abc xyz"); got != "xyzxyz\n" {
		t.Errorf("tr translate = %q", got)
	}
	if got := mustExec(t, fs, "echo abcabc | tr -d b"); got != "acac\n" {
		t.Errorf("tr delete = %q", got)
	}
}

func TestShufEchoDeterministic(t *testing.T) {
	fs := afero.NewMemMapFs()
	out := mustExec(t, fs, "shuf -e --seed 1 a b c")
	lines := strings.Fields(out)
	if len(lines) != 3 {
		t.Errorf("shuf -e produced %d lines: %q", len(lines), out)
	}
}

// findTree seeds a filesystem with a small, mixed file/directory tree rooted at
// /r for exercising find's filters at every depth.
func findTree(t *testing.T) afero.Fs {
	t.Helper()
	fs := afero.NewMemMapFs()
	for _, p := range []string{"/r/a.go", "/r/b.txt", "/r/sub/c.go", "/r/sub/deep/d.go"} {
		if err := afero.WriteFile(fs, p, []byte("x"), 0o644); err != nil {
			t.Fatalf("seed %q: %v", p, err)
		}
	}
	return fs
}

// TestFindBehavior drives the engine with real find flags over an in-memory tree
// and asserts the exact set of entries emitted. afero.Walk visits entries in
// lexical order, so the output is deterministic. This covers the neutral-default
// translation in buildFind: an unset name/type imposes no restriction and the
// default depth of -1 walks the whole tree, while -maxdepth 0 keeps only the
// root.
func TestFindBehavior(t *testing.T) {
	fs := findTree(t)
	cases := []struct {
		line string
		want string
	}{
		{"find /r", "/r\n/r/a.go\n/r/b.txt\n/r/sub\n/r/sub/c.go\n/r/sub/deep\n/r/sub/deep/d.go\n"},
		{"find /r --type f", "/r/a.go\n/r/b.txt\n/r/sub/c.go\n/r/sub/deep/d.go\n"},
		{"find /r --type d", "/r\n/r/sub\n/r/sub/deep\n"},
		{"find /r --maxdepth 0", "/r\n"},
		{"find /r --maxdepth 1", "/r\n/r/a.go\n/r/b.txt\n/r/sub\n"},
		{"find /r --name *.go", "/r/a.go\n/r/sub/c.go\n/r/sub/deep/d.go\n"},
		{"find /r --name *.go --type f --maxdepth 1", "/r/a.go\n"},
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("%q = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestNlBodyValues(t *testing.T) {
	for _, v := range []Argument{"a", "t", "n"} {
		if _, err := nlBody(v); err != nil {
			t.Errorf("nlBody(%q): %v", v, err)
		}
	}
	if _, err := nlBody("bad"); err == nil {
		t.Error("nlBody(bad) should fail")
	}
}
