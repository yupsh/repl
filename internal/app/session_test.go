package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/expansion"
	"github.com/yupsh/repl/internal/token"
)

// testHome is the injected home directory used by the session tests.
const testHome expansion.Home = "/home/tester"

// newSession builds a Session over an in-memory filesystem with captured output.
func newSession(input string, fs afero.Fs) (Session, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return New(strings.NewReader(input), out, errOut, fs, testHome), out, errOut
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
	eng, out, errOut := newSession(script, fs)
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
// way a *term.Terminal feeds the session in interactive mode.
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
	eng, out, _ := newSession("quit\n", fs)
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
	eng, out, _ := newSession("echo bye\n", fs)
	if err := eng.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "bye") {
		t.Errorf("output missing %q", "bye")
	}
}

// failingReader always errors, exercising the scanner-error return path.
type failingReader struct{}

var errReadFailed = errors.New("read failed")

func (failingReader) Read([]byte) (int, error) { return 0, errReadFailed }

func TestRunScannerError(t *testing.T) {
	fs := afero.NewMemMapFs()
	eng := New(failingReader{}, &bytes.Buffer{}, &bytes.Buffer{}, fs, testHome)
	if err := eng.Run(context.Background()); err == nil {
		t.Error("expected scanner error")
	}
}

func TestFirstWord(t *testing.T) {
	if got := firstWord(token.Line("help")); got != "help" {
		t.Errorf("firstWord = %q", got)
	}
	if got := firstWord(token.Line("echo hi")); got != "echo" {
		t.Errorf("firstWord = %q", got)
	}
}
