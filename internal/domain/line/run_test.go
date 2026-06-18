package line

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/token"
)

// testHome is the injected home directory used by tests for tilde expansion.
const testHome = "/home/tester"

// execLine plans and runs a single line against fs with the given stdin,
// returning the streamed output and any error.
func execLine(t *testing.T, fs afero.Fs, stdin, in string) (string, error) {
	t.Helper()
	cfg := Config{Fs: fs, Home: testHome}
	asm, err := Run(cfg, strings.NewReader(stdin), token.Line(in))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = asm.Run(context.Background(), &out)
	return out.String(), err
}

// mustExec runs a line, failing the test on error.
func mustExec(t *testing.T, fs afero.Fs, in string) string {
	t.Helper()
	out, err := execLine(t, fs, "", in)
	if err != nil {
		t.Fatalf("Run(%q): %v", in, err)
	}
	return out
}

func TestRunPipelines(t *testing.T) {
	fs := afero.NewMemMapFs()
	cases := []struct {
		line string
		want string
	}{
		{"echo hello world", "hello world\n"},
		{"seq 1 3", "1\n2\n3\n"},
		{"seq 1 5 | tac", "5\n4\n3\n2\n1\n"},
		{"seq 1 10 | grep 5", "5\n"},
		{"echo HELLO | tr A-Z a-z", "hello\n"},
		{"seq 1 100 | wc -l", "100\n"},
		{"emit hi there", "hi there\n"},
		{"seq 1 4 | head -n 2", "1\n2\n"},
		{"seq 1 4 | head -2", "1\n2\n"},
		{"seq 1 4 | tail -n 1", "4\n"},
		{"seq 1 4 | tail -1", "4\n"},
		{"comm -1 -2 -3", ""},
		{"echo abc | rev", "cba\n"},
		{"echo foo,bar | cut -d , -f 2", "bar\n"},
		{"basename /path/to/file.txt", "file.txt\n"},
		{"basename -s .txt /path/to/file.txt", "file\n"},
		{"dirname /path/to/file.txt", "/path/to\n"},
		{"emit /x/y | basename", "y\n"},
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("Run(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestRunXargsExec(t *testing.T) {
	fs := afero.NewMemMapFs()
	cases := []struct {
		line string
		want string
	}{
		// A command after xargs runs once per argument group, via the registry.
		{"echo a b c | xargs echo", "a b c\n"},
		{"seq 1 4 | xargs -n2 echo", "1 2\n3 4\n"},
		// -I substitutes the token per input line.
		{"seq 1 2 | xargs -I X echo row X", "row 1\nrow 2\n"},
		// With no command, xargs still regroups fields into argument lines.
		{"echo 1 2 3 4 | xargs -n2", "1 2\n3 4\n"},
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("Run(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestRunFileSource(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, "/data.txt", []byte("alpha\nbeta\ngamma\nbeta\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := mustExec(t, fs, "wc -l /data.txt"); got != "4\n" {
		t.Errorf("wc -l file = %q", got)
	}
	if got := mustExec(t, fs, "cat /data.txt | grep beta"); got != "beta\nbeta\n" {
		t.Errorf("cat|grep = %q", got)
	}
	if got := mustExec(t, fs, "sort /data.txt | uniq"); got != "alpha\nbeta\ngamma\n" {
		t.Errorf("sort|uniq = %q", got)
	}
}

func TestRunStdinSource(t *testing.T) {
	fs := afero.NewMemMapFs()
	out, err := execLine(t, fs, "one\ntwo\nthree\n", "grep two")
	if err != nil {
		t.Fatal(err)
	}
	if out != "two\n" {
		t.Errorf("stdin grep = %q", out)
	}
}

func TestRunGlobExpansion(t *testing.T) {
	fs := afero.NewMemMapFs()
	for path, body := range map[string]string{
		"/p/a.go":      "1\n2\n",
		"/p/b.go":      "3\n",
		"/p/readme.md": "x\n",
	} {
		if err := afero.WriteFile(fs, path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		line string
		want string
	}{
		{"wc -l /p/*.go", "3\n"},             // glob feeds files to a filter
		{"ls /p/*.go", "/p/a.go\n/p/b.go\n"}, // glob → names
		{"ls '/p/*.go'", "/p/*.go\n"},        // quoted → literal, no expansion
		{"echo /p/*.xml", "/p/*.xml\n"},      // no match → literal
		{"echo /p/[bad", "/p/[bad\n"},        // malformed pattern → literal
		{"ls /p", "a.go\nb.go\nreadme.md\n"}, // single directory → list entries
		{"ls /p/a.go", "/p/a.go\n"},          // single non-directory → echo name
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("Run(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestRunTildeExpansion(t *testing.T) {
	fs := afero.NewMemMapFs()
	if err := afero.WriteFile(fs, testHome+"/data.txt", []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := mustExec(t, fs, "wc -l ~/data.txt"); got != "3\n" {
		t.Errorf("wc -l ~/data.txt = %q, want %q", got, "3\n")
	}
}

func TestRunErrors(t *testing.T) {
	fs := afero.NewMemMapFs()
	cases := []struct {
		line    string
		wantErr constants.Error
	}{
		{"", constants.ErrEmptyCommand},
		{"bogus", constants.ErrUnknownCommand},
		{"wc -z", constants.ErrUnknownFlag},
		{"grep", constants.ErrMissingArgument},
		{"echo 'unterminated", constants.ErrUnterminatedQuote},
		{"seq one", constants.ErrInvalidNumber},
		{"seq 1 5 | echo hi", constants.ErrSourceMidPipeline},
		{"cat | wc /tmp/x.txt", constants.ErrArgsMidPipeline},
		{"seq 1 2 | basename /a/b", constants.ErrArgsMidPipeline},
	}
	for _, c := range cases {
		_, err := execLine(t, fs, "", c.line)
		if err == nil || !strings.Contains(err.Error(), string(c.wantErr)) {
			t.Errorf("Run(%q) err = %v, want %q", c.line, err, c.wantErr)
		}
	}
}

func TestRunRuntimeError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// A malformed sed expression fails per line, surfacing as a run error rather
	// than a planning error.
	if _, err := execLine(t, fs, "", "echo x | sed nonsense"); err == nil {
		t.Error("expected runtime error from malformed sed expression")
	}
}
