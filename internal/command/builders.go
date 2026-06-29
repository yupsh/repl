package command

import (
	"strings"

	echo "github.com/gloo-foo/cmd-echo"
	emit "github.com/gloo-foo/cmd-emit"
	find "github.com/gloo-foo/cmd-find"
	grep "github.com/gloo-foo/cmd-grep"
	ls "github.com/gloo-foo/cmd-ls"
	nl "github.com/gloo-foo/cmd-nl"
	sed "github.com/gloo-foo/cmd-sed"
	seq "github.com/gloo-foo/cmd-seq"
	shuf "github.com/gloo-foo/cmd-shuf"
	tr "github.com/gloo-foo/cmd-tr"
	yes "github.com/gloo-foo/cmd-yes"
	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/constants"
	"github.com/yupsh/repl/internal/flags"
)

// filterMaker builds a transform command from its parsed options.
type filterMaker func(opts []any) gloo.Command[[]byte, []byte]

// commandMaker builds a transform command from positionals and options, used by
// commands that consume their own positional arguments (exec, git, perl).
type commandMaker func(positional flags.Argv, opts []any) gloo.Command[[]byte, []byte]

// filter adapts a filterMaker into a BuildFunc, routing positional arguments to
// the pipeline source as files.
func filter(maker filterMaker) BuildFunc {
	return func(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
		return Segment{Command: maker(opts), Files: toFiles(positional)}, nil
	}
}

// command adapts a commandMaker into a BuildFunc; the command reads the pipeline
// input directly, so positionals are not routed to a file source.
func command(maker commandMaker) BuildFunc {
	return func(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
		return Segment{Command: maker(positional, opts)}, nil
	}
}

// literal adapts a filterMaker into a BuildFunc whose positional arguments are
// the command's input lines rather than filenames — used by path processors
// (basename, dirname) that transform their arguments as data.
func literal(maker filterMaker) BuildFunc {
	return func(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
		return Segment{Command: maker(opts), Inputs: toLines(positional)}, nil
	}
}

// --- named builders for commands with positional or value-dependent shapes ---

// buildEcho builds the echo source, joining arguments with spaces.
func buildEcho(_ afero.Fs, positional flags.Argv, _ []any) (Segment, error) {
	return Segment{Source: echo.Echo(positional.Strings()...)}, nil
}

// buildEmit builds the emit source from joined literal text.
func buildEmit(_ afero.Fs, positional flags.Argv, _ []any) (Segment, error) {
	return Segment{Source: emit.Emit(emit.EmitStdout(strings.Join(positional.Strings(), " ")))}, nil
}

// buildYes builds the yes source, optionally repeating a custom line.
func buildYes(_ afero.Fs, positional flags.Argv, _ []any) (Segment, error) {
	if len(positional) == 0 {
		return Segment{Source: yes.Yes()}, nil
	}
	return Segment{Source: yes.Yes(yes.YesText(strings.Join(positional.Strings(), " ")))}, nil
}

// buildSeq builds the seq source from numeric positionals plus flag options.
func buildSeq(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	args := make([]any, 0, len(positional)+len(opts))
	for _, p := range positional {
		n, err := flags.NumArg(p)
		if err != nil {
			return Segment{}, err
		}
		args = append(args, n)
	}
	args = append(args, opts...)
	return Segment{Source: seq.Seq(args...)}, nil
}

// buildLs builds the ls source. With no argument, or a single directory
// argument, it lists that directory's entries (the cmd-ls behavior). With
// multiple arguments or a single non-directory — typically the result of a glob
// like "*.go" — it emits the named paths themselves, matching how a shell's ls
// reports matched files.
func buildLs(fs afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	if lsListsNames(fs, positional) {
		return Segment{Source: gloo.SliceSource(toLines(positional))}, nil
	}
	opts = append(opts, ls.LsFs{Fs: fs})
	return Segment{Source: ls.Ls(flags.FirstOr(positional, "."), opts...)}, nil
}

// lsListsNames reports whether ls should echo its arguments as names rather than
// list a directory: true for multiple arguments or a single non-directory.
func lsListsNames(fs afero.Fs, positional flags.Argv) bool {
	if len(positional) == 0 {
		return false
	}
	if len(positional) > 1 {
		return true
	}
	isDir, _ := afero.DirExists(fs, string(positional[0]))
	return !isDir
}

// findName, findType, and findDepth are repl-local markers carrying a parsed
// find flag across the registry's []any boundary. cmd-find's option types
// (namePattern, typeFilter, maxDepth) and its switchFlag interface are
// unexported, so a []any cannot be spread into find.Find directly; buildFind
// translates these markers into the exported find.Find* constructors instead.
type (
	findName  string
	findType  string
	findDepth int
)

// buildFind builds the find source, injecting the environment filesystem and
// translating the parsed find markers into cmd-find options. Unset flags use
// the neutral defaults cmd-find applies when a flag is omitted: an empty name
// pattern matches every entry, an empty type imposes no kind restriction, and a
// depth of -1 is the unlimited-walk sentinel (so a user-supplied -maxdepth 0,
// meaning root only, stays distinct from "no -maxdepth given").
func buildFind(fs afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	name, typ, depth := "", "", -1
	for _, o := range opts {
		switch v := o.(type) {
		case findName:
			name = string(v)
		case findType:
			typ = string(v)
		case findDepth:
			depth = int(v)
		}
	}
	src := find.Find(flags.FirstOr(positional, "."),
		find.FindFs(fs),
		find.FindName(name),
		find.FindType(typ),
		find.FindMaxDepth(depth))
	return Segment{Source: src}, nil
}

// buildGrep builds the grep filter; the first positional is the pattern and the
// remainder are input files.
func buildGrep(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	if len(positional) == 0 {
		return Segment{}, constants.ErrMissingArgument.With(nil, "grep: pattern")
	}
	cmd := grep.Grep(string(positional[0]), opts...)
	return Segment{Command: cmd, Files: toFiles(positional[1:])}, nil
}

// buildSed builds the sed filter; the first positional is the expression and the
// remainder are input files.
func buildSed(_ afero.Fs, positional flags.Argv, _ []any) (Segment, error) {
	if len(positional) == 0 {
		return Segment{}, constants.ErrMissingArgument.With(nil, "sed: expression")
	}
	return Segment{Command: sed.Sed(string(positional[0])), Files: toFiles(positional[1:])}, nil
}

// buildTr builds the tr filter. SET1 is required; SET2 is consumed only when the
// command is not in delete mode. Any trailing positionals are input files.
func buildTr(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	if len(positional) == 0 {
		return Segment{}, constants.ErrMissingArgument.With(nil, "tr: SET1")
	}
	from := string(positional[0])
	rest := positional[1:]
	to := ""
	if trWantsSet2(opts) && len(rest) > 0 {
		to = string(rest[0])
		rest = rest[1:]
	}
	return Segment{Command: tr.Tr(from, to, opts...), Files: toFiles(rest)}, nil
}

// trWantsSet2 reports whether a second character set applies: every mode except
// pure deletion consumes SET2.
func trWantsSet2(opts []any) bool {
	for _, o := range opts {
		if o == tr.TrDelete {
			return false
		}
	}
	return true
}

// shufEchoMarker is the parsed-flag stand-in for shuf -e. It is removed before
// construction and replaced with a ShufEcho built from the positionals.
type shufEchoMarker struct{}

// buildShuf builds the shuf filter. With -e the positionals become the input
// lines; otherwise they are input files.
func buildShuf(_ afero.Fs, positional flags.Argv, opts []any) (Segment, error) {
	clean, echo := extractEcho(opts)
	if echo {
		clean = append(clean, shuf.ShufEcho(positional.Strings()...))
		return Segment{Command: shuf.Shuf(clean...)}, nil
	}
	return Segment{Command: shuf.Shuf(clean...), Files: toFiles(positional)}, nil
}

// extractEcho removes the shuf -e marker from opts, reporting whether it was
// present.
func extractEcho(opts []any) (clean []any, echo bool) {
	clean = make([]any, 0, len(opts))
	for _, o := range opts {
		if _, ok := o.(shufEchoMarker); ok {
			echo = true
			continue
		}
		clean = append(clean, o)
	}
	return clean, echo
}

// nlBody maps an -b value ("a", "t", "n") to the matching nl body-numbering
// option.
func nlBody(value flags.Argument) (any, error) {
	switch value {
	case "a":
		return nl.NlBodyAll, nil
	case "t":
		return nl.NlBodyNonEmpty, nil
	case "n":
		return nl.NlBodyNone, nil
	}
	return nil, constants.ErrInvalidFlagValue.With(nil, "nl -b: "+string(value))
}
