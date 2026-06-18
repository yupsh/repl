package repl

import (
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"
)

// Version is the REPL release. Bumped to 0.2.0 for the rework onto the
// gloo-foo/framework typed-stream API and the gloo-foo/cmd-* command modules.
const Version = "0.2.0"

// Line is one line of input typed at the prompt.
type Line string

// HomeDir is the home directory used for "~" expansion. An empty value disables
// tilde expansion.
type HomeDir string

// CommandName is a command keyword typed at the prompt (e.g. "grep").
type CommandName string

// Argument is a single parsed token of a command segment.
type Argument string

// Argv is the ordered argument list of one command segment.
type Argv []Argument

// environment carries the injected collaborators a builder or the expansion
// pass may need: the filesystem (for file, ls, and find construction and for
// glob matching) and the home directory (for tilde expansion). Both are injected
// so expansion and construction stay testable against an in-memory afero.Fs.
type environment struct {
	fs   afero.Fs
	home string
}

// segment is the built form of one pipeline stage. A source segment sets source
// and no command. A filter segment sets command and, optionally on the first
// stage, an input source: files are opened from the filesystem, while inputs
// are used verbatim as input lines (for path processors like basename whose
// arguments are data, not filenames). At most one of files or inputs is set.
type segment struct {
	source  gloo.Source[[]byte]
	command gloo.Command[[]byte, []byte]
	files   []gloo.File
	inputs  [][]byte
}

// buildFunc constructs a segment from a command's positional arguments and the
// already-parsed flag options.
type buildFunc func(env environment, positional Argv, opts []any) (segment, error)

// builder is a registry entry: how to parse a command's flags, how to build its
// segment, and a one-line summary for help output. When raw is set, flag
// parsing is skipped and every token is passed through as a positional — used
// by commands whose arguments are literal text (echo, yes, emit) or belong to
// an external program (git, exec).
type builder struct {
	flags   flagSet
	build   buildFunc
	raw     bool
	summary string
}

// toFiles converts positional arguments to framework File values.
func toFiles(positional Argv) []gloo.File {
	files := make([]gloo.File, len(positional))
	for i, a := range positional {
		files[i] = gloo.File(a)
	}
	return files
}

// toLines converts positional arguments to input lines, or nil when there are
// none (so the command falls back to stdin or its upstream stage).
func toLines(positional Argv) [][]byte {
	if len(positional) == 0 {
		return nil
	}
	lines := make([][]byte, len(positional))
	for i, a := range positional {
		lines[i] = []byte(a)
	}
	return lines
}

// strings converts positional arguments to a plain string slice.
func (a Argv) strings() []string {
	out := make([]string, len(a))
	for i, arg := range a {
		out[i] = string(arg)
	}
	return out
}

// anys converts positional arguments to an []any (for variadic ...any
// constructors that classify their inputs, such as Exec).
func (a Argv) anys() []any {
	out := make([]any, len(a))
	for i, arg := range a {
		out[i] = string(arg)
	}
	return out
}
