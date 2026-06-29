// Command yupsh is an interactive REPL over the gloo-foo/framework typed-stream
// pipeline and the gloo-foo/cmd-* command modules.
//
// It is a thin wiring shim: it binds the process's real stdin, stdout, stderr,
// and filesystem to the testable app.Session and forwards OS interrupts as
// context cancellation. On a terminal it drives the session through a
// golang.org/x/term.Terminal for line editing and command history; otherwise it
// falls back to plain line scanning. All behaviour lives in the internal/...
// packages, which are covered to 100%.
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

	"github.com/yupsh/repl/internal/app"
	"github.com/yupsh/repl/internal/expansion"
)

// version is the binary's version string. It defaults to "dev" for local
// builds and is overridden at release time via the linker:
// -ldflags "-X main.version=<version>" (set by goreleaser).
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Printf("yupsh version %s\n", version)
		return
	}
	if err := execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "yupsh: %v\n", err)
		os.Exit(1)
	}
}

// execute wires the process resources to the session and runs the REPL. It is
// split from main so the signal-context cleanup runs via defer before main
// calls os.Exit on failure.
func execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// A failed home lookup is non-fatal: tilde expansion simply stays literal.
	home, _ := os.UserHomeDir()
	return run(ctx, afero.NewOsFs(), expansion.Home(home))
}

// run selects an interactive terminal session when stdin is a TTY, and a plain
// scanning session otherwise (pipes, redirects, tests).
func run(ctx context.Context, fs afero.Fs, home expansion.Home) error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return app.New(os.Stdin, os.Stdout, os.Stderr, fs, home).Run(ctx)
	}
	return runInteractive(ctx, fd, fs, home)
}

// runInteractive puts the terminal into raw mode and drives the session through
// a term.Terminal, which supplies in-line editing and up/down history. All
// session output is routed through the terminal so raw-mode line endings render
// correctly.
func runInteractive(ctx context.Context, fd int, fs afero.Fs, home expansion.Home) error {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(fd, state) }()

	t := term.NewTerminal(readWriter{os.Stdin, os.Stdout}, app.Prompt)
	return app.NewWithReader(t, os.Stdin, t, t, fs, home).Run(ctx)
}

// readWriter joins a reader and a writer into the io.ReadWriter that
// term.NewTerminal expects.
type readWriter struct {
	io.Reader
	io.Writer
}
