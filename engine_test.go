package repl

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// testHome is the injected home directory used by tests for tilde expansion.
const testHome = "/home/tester"

// newEngine builds an Engine over an in-memory filesystem with captured output.
func newEngine(input string, fs afero.Fs) (Engine, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return New(strings.NewReader(input), out, errOut, fs, testHome), out, errOut
}

// execLine runs a single pipeline line and returns its clean output and error.
func execLine(t *testing.T, fs afero.Fs, line string) (string, error) {
	t.Helper()
	eng, out, _ := newEngine("", fs)
	err := eng.execute(context.Background(), Line(line))
	return out.String(), err
}

// mustExec runs a line, failing the test on error.
func mustExec(t *testing.T, fs afero.Fs, line string) string {
	t.Helper()
	out, err := execLine(t, fs, line)
	if err != nil {
		t.Fatalf("execute(%q): %v", line, err)
	}
	return out
}

func TestExecutePipelines(t *testing.T) {
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
		{"seq 1 3 | nl", "     1\t1\n     2\t2\n     3\t3\n"},
		{"echo foo,bar | cut -d , -f 2", "bar\n"},
		{"seq 1 3 | cat -n", "     1\t1\n     2\t2\n     3\t3\n"},
		{"basename /path/to/file.txt", "file.txt\n"},
		{"basename -s .txt /path/to/file.txt", "file\n"},
		{"dirname /path/to/file.txt", "/path/to\n"},
		{"emit /x/y | basename", "y\n"},
	}
	for _, c := range cases {
		if got := mustExec(t, fs, c.line); got != c.want {
			t.Errorf("execute(%q) = %q, want %q", c.line, got, c.want)
		}
	}
}

func TestExecuteFileSource(t *testing.T) {
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

func TestExecuteStdinSource(t *testing.T) {
	fs := afero.NewMemMapFs()
	out := &bytes.Buffer{}
	eng := New(strings.NewReader("one\ntwo\nthree\n"), out, &bytes.Buffer{}, fs, testHome)
	if err := eng.execute(context.Background(), Line("grep two")); err != nil {
		t.Fatal(err)
	}
	if out.String() != "two\n" {
		t.Errorf("stdin grep = %q", out.String())
	}
}

func TestExecuteErrors(t *testing.T) {
	fs := afero.NewMemMapFs()
	cases := []struct {
		line    string
		wantErr Error
	}{
		{"", ErrEmptyCommand},
		{"bogus", ErrUnknownCommand},
		{"wc -z", ErrUnknownFlag},
		{"grep", ErrMissingArgument},
		{"echo 'unterminated", ErrUnterminatedQuote},
		{"seq one", ErrInvalidNumber},
		{"seq 1 5 | echo hi", ErrSourceMidPipeline},
		{"cat | wc /tmp/x.txt", ErrArgsMidPipeline},
		{"seq 1 2 | basename /a/b", ErrArgsMidPipeline},
	}
	for _, c := range cases {
		_, err := execLine(t, fs, c.line)
		if err == nil || !strings.Contains(err.Error(), string(c.wantErr)) {
			t.Errorf("execute(%q) err = %v, want %q", c.line, err, c.wantErr)
		}
	}
}

func TestExecuteRuntimeError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// sed with a malformed expression fails per line, surfacing as a pipeline
	// run error rather than a build error.
	if _, err := execLine(t, fs, "echo x | sed nonsense"); err == nil {
		t.Error("expected runtime error from malformed sed expression")
	}
}

func TestRunLoopBuiltins(t *testing.T) {
	fs := afero.NewMemMapFs()
	script := strings.Join([]string{
		"",          // skipped blank
		"# comment", // skipped comment
		"help",      // built-in
		"version",   // built-in
		"clear",     // built-in
		"echo done", // command
		"bogus",     // error to stderr
		"exit",      // stop
		"never run", // after exit
	}, "\n") + "\n"
	eng, out, errOut := newEngine(script, fs)
	if err := eng.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	text := out.String()
	for _, want := range []string{"yupsh REPL v" + Version, "Built-ins:", "done", "Goodbye!"} {
		if !strings.Contains(text, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if !strings.Contains(errOut.String(), "unknown command") {
		t.Errorf("stderr missing error, got %q", errOut.String())
	}
	if strings.Contains(text, "never run") {
		t.Error("ran command after exit")
	}
}

// stubReader is an injected LineReader that replays a fixed list of lines, the
// way a *term.Terminal feeds the engine in interactive mode.
type stubReader struct {
	lines []string
	index int
}

func (r *stubReader) ReadLine() (string, error) {
	if r.index >= len(r.lines) {
		return "", io.EOF
	}
	line := r.lines[r.index]
	r.index++
	return line, nil
}

func TestNewWithReader(t *testing.T) {
	fs := afero.NewMemMapFs()
	out := &bytes.Buffer{}
	reader := &stubReader{lines: []string{"echo injected", "seq 1 2"}}
	eng := NewWithReader(reader, strings.NewReader(""), out, &bytes.Buffer{}, fs, testHome)
	if err := eng.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "injected") {
		t.Errorf("output missing injected line: %q", out.String())
	}
}

func TestRunStopsOnQuit(t *testing.T) {
	fs := afero.NewMemMapFs()
	eng, out, _ := newEngine("quit\n", fs)
	if err := eng.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Goodbye!") {
		t.Error("quit did not stop")
	}
}

func TestRunEndsAtEOF(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Input ends without an exit built-in: the scanner reader returns io.EOF and
	// the loop stops cleanly.
	eng, out, _ := newEngine("echo bye\n", fs)
	if err := eng.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "bye") {
		t.Errorf("output missing %q", "bye")
	}
}

func TestRunScannerError(t *testing.T) {
	fs := afero.NewMemMapFs()
	eng := New(failingReader{}, &bytes.Buffer{}, &bytes.Buffer{}, fs, testHome)
	if err := eng.Run(context.Background()); err == nil {
		t.Error("expected scanner error")
	}
}

// failingReader always errors, exercising the scanner-error return path.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errReadFailed }

const errReadFailed Error = "read failed"

func TestFirstWord(t *testing.T) {
	if got := firstWord("help"); got != "help" {
		t.Errorf("firstWord = %q", got)
	}
	if got := firstWord("echo hi"); got != "echo" {
		t.Errorf("firstWord = %q", got)
	}
}
