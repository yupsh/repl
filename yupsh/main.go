// Command yupsh is an interactive REPL over the gloo-foo/framework typed-stream
// pipeline and the gloo-foo/cmd-* command modules.
//
// It is a thin wiring shim: it binds the process's real stdin, stdout, stderr,
// and filesystem to the testable repl.Engine and forwards OS interrupts as
// context cancellation. On a terminal it drives the engine through a
// golang.org/x/term.Terminal for line editing and command history; otherwise it
// falls back to plain line scanning. All behaviour lives in the
// github.com/yupsh/repl package, which is covered to 100%.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/afero"
	"golang.org/x/term"

	"github.com/yupsh/repl"
)

// appVersion is the binary's version string. It defaults to "dev" for local
// builds and is overridden at release time via the linker:
// -ldflags "-X main.appVersion=<version>" (set by goreleaser).
var appVersion = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("yupsh version %s\n", appVersion)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// A failed home lookup is non-fatal: tilde expansion simply stays literal.
	home, _ := os.UserHomeDir()
	if err := run(ctx, afero.NewOsFs(), repl.HomeDir(home)); err != nil {
		fmt.Fprintf(os.Stderr, "yupsh: %v\n", err)
		os.Exit(1)
	}
}

// run selects an interactive terminal session when stdin is a TTY, and a plain
// scanning session otherwise (pipes, redirects, tests).
func run(ctx context.Context, fs afero.Fs, home repl.HomeDir) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return repl.New(os.Stdin, os.Stdout, os.Stderr, fs, home).Run(ctx)
	}
	return runInteractive(ctx, fd, fs, home)
}

// runInteractive puts the terminal into raw mode and drives the engine through
// a term.Terminal, which supplies in-line editing and up/down history. All
// engine output is routed through the terminal so raw-mode line endings render
// correctly.
func runInteractive(ctx context.Context, fd int, fs afero.Fs, home repl.HomeDir) error {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(fd, state) }()

	t := term.NewTerminal(readWriter{os.Stdin, os.Stdout}, repl.Prompt)
	return repl.NewWithReader(t, os.Stdin, t, t, fs, home).Run(ctx)
}

// readWriter joins a reader and a writer into the io.ReadWriter that
// term.NewTerminal expects.
type readWriter struct {
	io.Reader
	io.Writer
}
