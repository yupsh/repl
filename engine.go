package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"
)

// Engine is an immutable REPL instance wired to injected collaborators: a
// LineReader for command input, an io.Reader for pipeline stdin, output and
// error writers, and a filesystem. Everything is injected so the engine is
// fully testable with in-memory buffers and an afero.MemMapFs. Value receiver on
// every method — no mutation.
type Engine struct {
	reader LineReader
	stdin  io.Reader
	out    io.Writer
	errw   io.Writer
	env    environment
	reg    map[CommandName]builder
}

// New builds an Engine that reads command lines from in (with no terminal
// editing) and uses in as the pipeline stdin source, writing results to out and
// errors to errw. File arguments and globs resolve against fs, and "~" expands
// to home.
func New(in io.Reader, out, errw io.Writer, fs afero.Fs, home HomeDir) Engine {
	return NewWithReader(newScanReader(in, out), in, out, errw, fs, home)
}

// NewWithReader builds an Engine whose command lines come from an explicit
// LineReader (e.g. a *golang.org/x/term.Terminal for interactive history and
// editing), with stdin as the pipeline input source.
func NewWithReader(reader LineReader, stdin io.Reader, out, errw io.Writer, fs afero.Fs, home HomeDir) Engine {
	return Engine{
		reader: reader,
		stdin:  stdin,
		out:    out,
		errw:   errw,
		env:    environment{fs: fs, home: string(home)},
		reg:    registry(),
	}
}

// Run executes the read-eval-print loop until end of input or an exit built-in,
// returning any input error other than io.EOF.
func (e Engine) Run(ctx context.Context) error {
	e.banner()
	for {
		raw, err := e.reader.ReadLine()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if e.dispatch(ctx, Line(strings.TrimSpace(raw))) {
			return nil
		}
	}
}

// dispatch handles one input line, returning true when the loop should stop.
// Blank lines and "#" comments are ignored.
func (e Engine) dispatch(ctx context.Context, line Line) (stop bool) {
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
func (e Engine) builtin(line Line) (handled, stop bool) {
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

// execute parses, builds, and runs one pipeline line.
func (e Engine) execute(ctx context.Context, line Line) error {
	segments, err := parseLine(line)
	if err != nil {
		return err
	}
	stages, err := e.buildStages(segments)
	if err != nil {
		return err
	}
	a, err := plan(stages, e.env, e.stdin)
	if err != nil {
		return err
	}
	return a.run(ctx, e.out)
}

// buildStages builds every pipeline segment.
func (e Engine) buildStages(segments []rawSegment) ([]stage, error) {
	stages := make([]stage, 0, len(segments))
	for _, s := range segments {
		st, err := e.buildStage(s)
		if err != nil {
			return nil, err
		}
		stages = append(stages, st)
	}
	return stages, nil
}

// buildStage resolves and builds a single segment.
func (e Engine) buildStage(s rawSegment) (stage, error) {
	if len(s) == 0 {
		return stage{}, ErrEmptyCommand
	}
	name := CommandName(s[0].text)
	b, ok := e.reg[name]
	if !ok {
		return stage{}, ErrUnknownCommand.With(nil, string(name))
	}
	opts, positional, err := resolveArgs(b, expandArgs(e.env, s[1:]))
	if err != nil {
		return stage{}, err
	}
	seg, err := b.build(e.env, positional, opts)
	if err != nil {
		return stage{}, err
	}
	return stage{name: name, segment: seg}, nil
}

// resolveArgs splits a segment's arguments into options and positionals,
// honouring a builder's raw passthrough mode.
func resolveArgs(b builder, args Argv) (opts []any, positional Argv, err error) {
	if b.raw {
		return nil, args, nil
	}
	return parseArgs(b.flags, args)
}

// banner prints the startup header.
func (e Engine) banner() {
	fmt.Fprintf(e.out, "yupsh REPL v%s\n", Version)
	fmt.Fprintln(e.out, "Yup Shell — a typed-stream REPL over gloo-foo/framework and cmd-* commands.")
	fmt.Fprintln(e.out, "Type 'help' for commands, 'exit' to quit.")
	fmt.Fprintln(e.out)
}

// firstWord returns the leading whitespace-delimited word of a line.
func firstWord(line Line) string {
	s := string(line)
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}
