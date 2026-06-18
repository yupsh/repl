//go:build integration

// Package repl_test holds black-box integration tests for the yupsh binary.
//
// These tests build the real `yupsh` executable and drive it through stdin,
// asserting on its observable output — the contract a user relies on. They are
// gated behind the `integration` build tag (run with `go test -tags integration
// ./...`) so they are opt-in and never part of the hermetic unit gate.
//
// Scope: they verify the capabilities THIS repository implements — the shell
// front-end — not the behavior of the underlying gloo-foo/cmd-* commands. So we
// assert that `seq`/`echo` source, that pipes compose, that a Unix flag is
// *translated* and reaches its command, that globs and `~` expand, that files
// and quoting work, and that built-ins and errors behave. We do NOT re-test that
// `wc` counts correctly or that `grep` matches correctly — those are each
// command's own responsibility, covered in their repositories.
package repl_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binPath is the compiled yupsh binary, built once in TestMain.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "yupsh-it-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	binPath = filepath.Join(dir, "yupsh")
	if out, err := exec.Command("go", "build", "-o", binPath, "./yupsh").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build yupsh: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// session is a sandbox: a working directory and a home directory, both real and
// temporary, into which tests place files and against which the binary runs.
type session struct {
	dir  string
	home string
}

func newSession(t *testing.T) session {
	t.Helper()
	return session{dir: t.TempDir(), home: t.TempDir()}
}

// write creates a file under the working directory.
func (s session) write(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(s.dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeHome creates a file under the home directory.
func (s session) writeHome(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(s.home, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// run feeds script to the binary's stdin and returns cleaned command-output
// lines (banner and prompts stripped) plus raw stderr.
func (s session) run(t *testing.T, script string) (out []string, stderr string) {
	t.Helper()
	cmd := exec.Command(binPath)
	cmd.Dir = s.dir
	cmd.Env = append(os.Environ(), "HOME="+s.home)
	cmd.Stdin = strings.NewReader(script)
	var o, e bytes.Buffer
	cmd.Stdout, cmd.Stderr = &o, &e
	_ = cmd.Run() // the REPL exits 0 on EOF/exit; command errors go to stderr
	return outputLines(o.String()), e.String()
}

// outputLines drops the four-line startup banner, strips the "yup> " prompt that
// precedes each line of command output, and discards the blank lines left by
// bare prompts — leaving just what commands wrote.
func outputLines(stdout string) []string {
	raw := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(raw) >= 4 {
		raw = raw[4:] // banner: version, subtitle, hint, blank
	}
	var out []string
	for _, ln := range raw {
		// A skipped line (blank/comment) emits a prompt with no output, so
		// several prompts can accumulate on one line — strip them all.
		for strings.HasPrefix(ln, "yup> ") {
			ln = ln[len("yup> "):]
		}
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

func wantLines(t *testing.T, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("output =\n%q\nwant\n%q", got, want)
	}
}

func wantHas(t *testing.T, got []string, line string) {
	t.Helper()
	for _, g := range got {
		if g == line {
			return
		}
	}
	t.Errorf("output %q does not contain line %q", got, line)
}

func wantAbsent(t *testing.T, got []string, line string) {
	t.Helper()
	for _, g := range got {
		if g == line {
			t.Errorf("output %q unexpectedly contains line %q", got, line)
		}
	}
}

// --- capability: sources and pipeline composition ---

func TestSourcesAndPipes(t *testing.T) {
	s := newSession(t)

	echoed, _ := s.run(t, "echo hello world\n")
	wantLines(t, echoed, "hello world")

	seqd, _ := s.run(t, "seq 1 3\n")
	wantLines(t, seqd, "1", "2", "3")

	// a three-stage pipeline composes: reverse 1..5 then take the first two.
	piped, _ := s.run(t, "seq 1 5 | tac | head -2\n")
	wantLines(t, piped, "5", "4")
}

// --- capability: flag translation reaches the command (one representative) ---

func TestFlagTranslation(t *testing.T) {
	s := newSession(t)
	// -v must be translated to grep's invert option: 3 is dropped.
	out, _ := s.run(t, "seq 1 4 | grep -v 3\n")
	wantLines(t, out, "1", "2", "4")
	// GNU "-NUM" shorthand must be translated to head's line count.
	out2, _ := s.run(t, "seq 1 9 | head -2\n")
	wantLines(t, out2, "1", "2")
}

// --- capability: file arguments on the first stage are read as input ---

func TestFileInput(t *testing.T) {
	s := newSession(t)
	s.write(t, "data.txt", "alpha\nbeta\ngamma\n")
	out, _ := s.run(t, "cat data.txt | grep beta\n")
	wantLines(t, out, "beta")
}

// --- capability: globbing ---

func TestGlobbing(t *testing.T) {
	s := newSession(t)
	s.write(t, "a.txt", "x\n")
	s.write(t, "b.txt", "y\n")
	s.write(t, "note.md", "z\n")

	out, _ := s.run(t, "ls *.txt\n")
	wantLines(t, out, "a.txt", "b.txt")

	// quoted glob is literal (suppressed expansion)
	q, _ := s.run(t, "echo '*.txt'\n")
	wantLines(t, q, "*.txt")

	// non-matching glob stays literal
	n, _ := s.run(t, "echo *.none\n")
	wantLines(t, n, "*.none")
}

// --- capability: tilde expansion ---

func TestTilde(t *testing.T) {
	s := newSession(t)
	s.writeHome(t, "h.txt", "one\ntwo\n")
	out, _ := s.run(t, "cat ~/h.txt\n")
	wantLines(t, out, "one", "two")
	echoed, _ := s.run(t, "echo ~\n")
	wantLines(t, echoed, s.home)
}

// --- capability: quoting groups arguments and preserves spacing ---

func TestQuoting(t *testing.T) {
	s := newSession(t)
	out, _ := s.run(t, "echo \"a  b\"\n")
	wantLines(t, out, "a  b")
}

// --- capability: built-ins and comments ---

func TestBuiltins(t *testing.T) {
	s := newSession(t)
	out, _ := s.run(t, "# a comment, ignored\nversion\nexit\necho after-exit\n")
	wantHas(t, out, "yupsh REPL v0.2.0")
	wantAbsent(t, out, "after-exit") // exit stops the loop

	help, _ := s.run(t, "help\n")
	wantHas(t, help, "Commands:")
}

// --- capability: error reporting ---

func TestErrors(t *testing.T) {
	s := newSession(t)
	cases := []struct {
		script  string
		wantErr string
	}{
		{"bogus\nexit\n", "unknown command"},
		{"seq 1 3 | echo hi\nexit\n", "source command cannot appear after a pipe"},
		{"echo \"unterminated\nexit\n", "unterminated quote"},
	}
	for _, c := range cases {
		if _, stderr := s.run(t, c.script); !strings.Contains(stderr, c.wantErr) {
			t.Errorf("script %q stderr = %q, want substring %q", c.script, stderr, c.wantErr)
		}
	}
}

// --- capability: exec wiring (real subprocess) ---

func TestExecWiring(t *testing.T) {
	s := newSession(t)
	out, _ := s.run(t, "echo piped | exec cat\n")
	wantLines(t, out, "piped")
}

// --- capability: perl wiring (real subprocess, skipped if perl is absent) ---

func TestPerlWiring(t *testing.T) {
	if _, err := exec.LookPath("perl"); err != nil {
		t.Skip("perl not installed")
	}
	s := newSession(t)
	out, _ := s.run(t, "echo hello | perl -p 's/l/L/g'\n")
	wantLines(t, out, "heLLo")
}
