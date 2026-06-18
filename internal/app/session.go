// Package app is the yupsh REPL surface: the interactive read-eval-print loop,
// its built-ins, banner, and help, plus the line reader that feeds it.
//
// It is the app tier — the CLI surface of the program. It owns no business
// logic: each input line is handed to the line domain (internal/domain/line),
// which plans a pipeline.Assembly, and the session "renders" that result by
// streaming it to the output writer. Built-ins (exit, help, version, clear) and
// I/O wiring live here; tokenizing, expansion, flag translation, and pipeline
// assembly do not.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"

	domain "github.com/yupsh/repl/internal/domain/line"
	"github.com/yupsh/repl/internal/expansion"
	"github.com/yupsh/repl/internal/token"
)

// Version is the REPL release. Bumped to 0.2.0 for the rework onto the
// gloo-foo/framework typed-stream API and the gloo-foo/cmd-* command modules.
const Version = "0.2.0"

// Session is an immutable REPL instance wired to injected collaborators: a
// LineReader for command input, an io.Reader for pipeline stdin, output and
// error writers, and the line domain's Config (filesystem + home). Everything is
// injected so the session is fully testable with in-memory buffers and an
// afero.MemMapFs. Value receiver on every method — no mutation.
type Session struct {
	reader LineReader
	stdin  io.Reader
	out    io.Writer
	errw   io.Writer
	cfg    domain.Config
}

// New builds a Session that reads command lines from in (with no terminal
// editing) and uses in as the pipeline stdin source, writing results to out and
// errors to errw. File arguments and globs resolve against fs, and "~" expands
// to home.
func New(in io.Reader, out, errw io.Writer, fs afero.Fs, home expansion.Home) Session {
	return NewWithReader(newScanReader(in, out), in, out, errw, fs, home)
}

// NewWithReader builds a Session whose command lines come from an explicit
// LineReader (e.g. a *golang.org/x/term.Terminal for interactive history and
// editing), with stdin as the pipeline input source.
func NewWithReader(reader LineReader, stdin io.Reader, out, errw io.Writer, fs afero.Fs, home expansion.Home) Session {
	return Session{
		reader: reader,
		stdin:  stdin,
		out:    out,
		errw:   errw,
		cfg:    domain.Config{Fs: fs, Home: home},
	}
}

// Run executes the read-eval-print loop until end of input or an exit built-in,
// returning any input error other than io.EOF.
func (e Session) Run(ctx context.Context) error {
	e.banner()
	for {
		raw, err := e.reader.ReadLine()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if e.dispatch(ctx, token.Line(strings.TrimSpace(raw))) {
			return nil
		}
	}
}

// dispatch handles one input line, returning true when the loop should stop.
// Blank lines and "#" comments are ignored.
func (e Session) dispatch(ctx context.Context, line token.Line) (stop bool) {
	if line == "" || strings.HasPrefix(string(line), "#") {
		return false
	}
	if handled, stop := e.builtin(line); handled {
		return stop
	}
	if err := e.execute(ctx, line); err != nil {
		fmt.Fprintf(e.errw, "yupsh: %v\n", err)
	}
	return false
}

// builtin handles shell built-ins, reporting whether the line was a built-in
// and whether the loop should stop.
func (e Session) builtin(line token.Line) (handled, stop bool) {
	switch firstWord(line) {
	case "exit", "quit":
		fmt.Fprintln(e.out, "Goodbye!")
		return true, true
	case "help":
		e.help()
		return true, false
	case "version":
		fmt.Fprintf(e.out, "yupsh REPL v%s\n", Version)
		return true, false
	case "clear":
		fmt.Fprint(e.out, "\033[H\033[2J")
		return true, false
	}
	return false, false
}

// execute plans one pipeline line via the line domain, then streams the
// assembled result to the output writer.
func (e Session) execute(ctx context.Context, line token.Line) error {
	asm, err := domain.Run(e.cfg, e.stdin, line)
	if err != nil {
		return err
	}
	return asm.Run(ctx, e.out)
}

// banner prints the startup header.
func (e Session) banner() {
	fmt.Fprintf(e.out, "yupsh REPL v%s\n", Version)
	fmt.Fprintln(e.out, "Yup Shell — a typed-stream REPL over gloo-foo/framework and cmd-* commands.")
	fmt.Fprintln(e.out, "Type 'help' for commands, 'exit' to quit.")
	fmt.Fprintln(e.out)
}

// firstWord returns the leading whitespace-delimited word of a line.
func firstWord(line token.Line) string {
	s := string(line)
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}
