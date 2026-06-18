// Package command is the yupsh command catalogue: the registry mapping each
// command keyword to how its flags are parsed and how its pipeline segment is
// built from the gloo-foo/cmd-* modules.
//
// Each registry entry is a Builder: a flag table (a flags.Set), a BuildFunc
// that turns parsed positionals and options into a Segment, an optional raw
// passthrough mode for commands whose arguments are literal text, and a one-line
// summary for help output. A Segment is the built form of one pipeline stage —
// a source, a transform command, and at most one input source (files or literal
// inputs). The package knows how to construct commands; it does not run them
// (that is internal/pipeline) nor orchestrate a line (that is the line domain).
package command

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/flags"
)

// Name is a command keyword typed at the prompt (e.g. "grep").
type Name string

// Segment is the built form of one pipeline stage. A source segment sets Source
// and no command. A filter segment sets Command and, optionally on the first
// stage, an input source: Files are opened from the filesystem, while Inputs are
// used verbatim as input lines (for path processors like basename whose
// arguments are data, not filenames). At most one of Files or Inputs is set.
type Segment struct {
	Source  gloo.Source[[]byte]
	Command gloo.Command[[]byte, []byte]
	Files   []gloo.File
	Inputs  [][]byte
}

// BuildFunc constructs a Segment from a command's positional arguments and the
// already-parsed flag options. The filesystem is injected for the commands (ls,
// find) that build against it.
type BuildFunc func(fs afero.Fs, positional flags.Argv, opts []any) (Segment, error)

// Builder is a registry entry: how to parse a command's flags, how to build its
// segment, and a one-line summary for help output. When Raw is set, flag parsing
// is skipped and every token is passed through as a positional — used by
// commands whose arguments are literal text (echo, yes, emit) or belong to an
// external program (git, exec).
type Builder struct {
	Flags   flags.Set
	Build   BuildFunc
	Raw     bool
	Summary string
}

// toFiles converts positional arguments to framework File values.
func toFiles(positional flags.Argv) []gloo.File {
	files := make([]gloo.File, len(positional))
	for i, a := range positional {
		files[i] = gloo.File(a)
	}
	return files
}

// toLines converts positional arguments to input lines, or nil when there are
// none (so the command falls back to stdin or its upstream stage).
func toLines(positional flags.Argv) [][]byte {
	if len(positional) == 0 {
		return nil
	}
	lines := make([][]byte, len(positional))
	for i, a := range positional {
		lines[i] = []byte(a)
	}
	return lines
}
