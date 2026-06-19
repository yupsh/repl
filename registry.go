package repl

import (
	"strings"

	gloo "github.com/gloo-foo/framework"
	"github.com/spf13/afero"

	base64 "github.com/gloo-foo/cmd-base64"
	basename "github.com/gloo-foo/cmd-basename"
	cat "github.com/gloo-foo/cmd-cat"
	comm "github.com/gloo-foo/cmd-comm"
	cut "github.com/gloo-foo/cmd-cut"
	diff "github.com/gloo-foo/cmd-diff"
	dirname "github.com/gloo-foo/cmd-dirname"
	echo "github.com/gloo-foo/cmd-echo"
	emit "github.com/gloo-foo/cmd-emit"
	exec "github.com/gloo-foo/cmd-exec"
	find "github.com/gloo-foo/cmd-find"
	git "github.com/gloo-foo/cmd-git"
	grep "github.com/gloo-foo/cmd-grep"
	head "github.com/gloo-foo/cmd-head"
	hexdump "github.com/gloo-foo/cmd-hexdump"
	join "github.com/gloo-foo/cmd-join"
	jsoncmd "github.com/gloo-foo/cmd-json"
	ls "github.com/gloo-foo/cmd-ls"
	nl "github.com/gloo-foo/cmd-nl"
	paste "github.com/gloo-foo/cmd-paste"
	perl "github.com/gloo-foo/cmd-perl"
	rev "github.com/gloo-foo/cmd-rev"
	sed "github.com/gloo-foo/cmd-sed"
	seq "github.com/gloo-foo/cmd-seq"
	shuf "github.com/gloo-foo/cmd-shuf"
	sortcmd "github.com/gloo-foo/cmd-sort"
	split "github.com/gloo-foo/cmd-split"
	tac "github.com/gloo-foo/cmd-tac"
	tail "github.com/gloo-foo/cmd-tail"
	tee "github.com/gloo-foo/cmd-tee"
	tr "github.com/gloo-foo/cmd-tr"
	uniq "github.com/gloo-foo/cmd-uniq"
	wc "github.com/gloo-foo/cmd-wc"
	xargs "github.com/gloo-foo/cmd-xargs"
	yes "github.com/gloo-foo/cmd-yes"
)

// registry returns the command table. It is built fresh per Engine so callers
// never share mutable state.
func registry() map[CommandName]builder {
	return map[CommandName]builder{
		"echo":     {raw: true, build: buildEcho, summary: "emit arguments as a line (source)"},
		"emit":     {raw: true, build: buildEmit, summary: "emit literal text (source)"},
		"yes":      {raw: true, build: buildYes, summary: "repeat a line until interrupted (source)"},
		"seq":      {flags: seqFlags, build: buildSeq, summary: "generate a numeric sequence (source)"},
		"ls":       {flags: lsFlags, build: buildLs, summary: "list directory entries (source)"},
		"find":     {flags: findFlags, build: buildFind, summary: "walk a directory tree (source)"},
		"base64":   {flags: base64Flags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return base64.Base64(o...) }), summary: "base64 encode or decode"},
		"basename": {flags: basenameFlags, build: literal(func(o []any) gloo.Command[[]byte, []byte] { return basename.Basename(o...) }), summary: "strip directory and suffix"},
		"cat":      {flags: catFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return cat.Cat(o...) }), summary: "concatenate with optional numbering"},
		"comm":     {flags: commFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return comm.Comm(o...) }), summary: "compare sorted streams"},
		"cut":      {flags: cutFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return cut.Cut(o...) }), summary: "select fields, bytes, or characters"},
		"diff":     {flags: diffFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return diff.Diff(o...) }), summary: "compare streams line by line"},
		"dirname":  {build: literal(func(o []any) gloo.Command[[]byte, []byte] { return dirname.Dirname(o...) }), summary: "strip last path component"},
		"head":     {flags: headFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return head.Head(o...) }), summary: "output the first lines or bytes"},
		"hexdump":  {flags: hexdumpFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return hexdump.Hexdump(o...) }), summary: "hex dump of input"},
		"join":     {flags: joinFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return join.Join(o...) }), summary: "join lines on a common field"},
		"json":     {build: filter(func(o []any) gloo.Command[[]byte, []byte] { return jsoncmd.Json(o...) }), summary: "pretty-print JSON"},
		"nl":       {flags: nlFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return nl.Nl(o...) }), summary: "number lines"},
		"paste":    {flags: pasteFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return paste.Paste(o...) }), summary: "merge lines"},
		"rev":      {build: filter(func(o []any) gloo.Command[[]byte, []byte] { return rev.Rev(o...) }), summary: "reverse each line"},
		"shuf":     {flags: shufFlags, build: buildShuf, summary: "shuffle lines randomly"},
		"sort":     {flags: sortFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return sortcmd.Sort(o...) }), summary: "sort lines"},
		"split":    {flags: splitFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return split.Split(o...) }), summary: "split fields onto lines"},
		"tac":      {flags: tacFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tac.Tac(o...) }), summary: "reverse line order"},
		"tail":     {flags: tailFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tail.Tail(o...) }), summary: "output the last lines or bytes"},
		"tee":      {build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tee.Tee(o...) }), summary: "pass input through unchanged"},
		"uniq":     {flags: uniqFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return uniq.Uniq(o...) }), summary: "drop adjacent duplicate lines"},
		"wc":       {flags: wcFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return wc.Wc(o...) }), summary: "count lines, words, and bytes"},
		"xargs":    {flags: xargsFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] { return xargs.Xargs(o...) }), summary: "group fields into argument lines"},
		"grep":     {flags: grepFlags, build: buildGrep, summary: "filter lines matching a pattern"},
		"sed":      {build: buildSed, summary: "apply an s/// substitution"},
		"tr":       {flags: trFlags, build: buildTr, summary: "translate, delete, or squeeze characters"},
		"exec":     {raw: true, build: command(func(p Argv, _ []any) gloo.Command[[]byte, []byte] { return exec.Exec(p.anys()...) }), summary: "run an external program"},
		"git":      {raw: true, build: command(func(p Argv, _ []any) gloo.Command[[]byte, []byte] { return git.Git(p.strings()...) }), summary: "run git"},
		"perl":     {flags: perlFlags, build: command(func(p Argv, o []any) gloo.Command[[]byte, []byte] { return perl.Perl(append(p.anys(), o...)...) }), summary: "run a perl one-liner"},
	}
}

// --- flag tables ---

var (
	base64Flags = flagSet{boolFlag("d", "decode", base64.Base64Decode)}

	basenameFlags = flagSet{valueFlag("s", "suffix", strMaker(func(s string) any { return basename.BasenameSuffix(s) }))}

	catFlags = flagSet{
		boolFlag("n", "number", cat.CatNumberLines),
		boolFlag("b", "number-nonblank", cat.CatNumberNonBlank),
	}

	commFlags = flagSet{
		boolFlag("1", "", comm.CommSuppressColumn1),
		boolFlag("2", "", comm.CommSuppressColumn2),
		boolFlag("3", "", comm.CommSuppressColumn3),
	}

	cutFlags = flagSet{
		valueFlag("d", "delimiter", strMaker(func(s string) any { return cut.CutDelimiter(s) })),
		valueFlag("f", "fields", func(v Argument) (any, error) {
			xs, err := intList(v)
			if err != nil {
				return nil, err
			}
			return cut.CutFields(xs...), nil
		}),
		valueFlag("b", "bytes", strMaker(func(s string) any { return cut.CutBytes(s) })),
		valueFlag("c", "characters", strMaker(func(s string) any { return cut.CutChars(s) })),
		boolFlag("", "complement", cut.CutComplement),
	}

	diffFlags = flagSet{boolFlag("u", "unified", diff.DiffUnified)}

	headFlags = flagSet{
		valueFlag("n", "lines", intMaker(func(n int) any { return head.HeadLines(n) })),
		valueFlag("c", "bytes", intMaker(func(n int) any { return head.HeadBytes(n) })),
		numFlag(intMaker(func(n int) any { return head.HeadLines(n) })),
	}

	hexdumpFlags = flagSet{boolFlag("C", "canonical", hexdump.HexdumpCanonical)}

	joinFlags = flagSet{valueFlag("t", "separator", strMaker(func(s string) any { return join.JoinSeparator(s) }))}

	nlFlags = flagSet{
		valueFlag("b", "body-numbering", nlBody),
		valueFlag("s", "separator", strMaker(func(s string) any { return nl.NlSep(s) })),
		valueFlag("v", "starting-line", intMaker(func(n int) any { return nl.NlStart(n) })),
		valueFlag("i", "increment", intMaker(func(n int) any { return nl.NlIncrement(n) })),
		valueFlag("w", "width", intMaker(func(n int) any { return nl.NlWidth(n) })),
		valueFlag("n", "number-format", strMaker(func(s string) any { return nl.NlFormat(s) })),
	}

	pasteFlags = flagSet{
		valueFlag("d", "delimiters", strMaker(func(s string) any { return paste.PasteDelimiter(s) })),
		boolFlag("s", "serial", paste.PasteSerial),
	}

	sortFlags = flagSet{
		boolFlag("r", "reverse", sortcmd.SortReverse),
		boolFlag("n", "numeric-sort", sortcmd.SortNumeric),
		boolFlag("u", "unique", sortcmd.SortUnique),
		boolFlag("f", "ignore-case", sortcmd.SortIgnoreCase),
		boolFlag("R", "random-sort", sortcmd.SortRandom),
		boolFlag("b", "ignore-leading-blanks", sortcmd.SortIgnoreLeadingBlanks),
		boolFlag("V", "version-sort", sortcmd.SortVersionSort),
		boolFlag("h", "human-numeric-sort", sortcmd.SortHumanNumeric),
		boolFlag("M", "month-sort", sortcmd.SortMonthSort),
		boolFlag("s", "stable", sortcmd.SortStableSort),
		valueFlag("k", "key", intMaker(func(n int) any { return sortcmd.SortField(n) })),
		valueFlag("t", "field-separator", strMaker(func(s string) any { return sortcmd.SortDelimiter(s) })),
	}

	splitFlags = flagSet{valueFlag("d", "delimiter", strMaker(func(s string) any { return split.SplitDelim(s) }))}

	tacFlags = flagSet{valueFlag("s", "separator", strMaker(func(s string) any { return tac.TacSep(s) }))}

	tailFlags = flagSet{
		valueFlag("n", "lines", intMaker(func(n int) any { return tail.TailLines(n) })),
		valueFlag("c", "bytes", intMaker(func(n int) any { return tail.TailBytes(n) })),
		numFlag(intMaker(func(n int) any { return tail.TailLines(n) })),
	}

	uniqFlags = flagSet{
		boolFlag("d", "repeated", uniq.UniqDuplicatesOnly),
		boolFlag("c", "count", uniq.UniqCount),
	}

	shufFlags = flagSet{
		valueFlag("n", "head-count", intMaker(func(n int) any { return shuf.ShufCount(n) })),
		valueFlag("", "seed", int64Maker(func(n int64) any { return shuf.ShufSeed(n) })),
		valueFlag("i", "input-range", func(v Argument) (any, error) {
			lo, hi, err := rangeArg(v)
			if err != nil {
				return nil, err
			}
			return shuf.ShufRange(lo, hi), nil
		}),
		boolFlag("e", "echo", shufEchoMarker{}),
	}

	wcFlags = flagSet{
		boolFlag("l", "lines", wc.WcLines),
		boolFlag("w", "words", wc.WcWords),
		boolFlag("c", "bytes", wc.WcBytes),
		boolFlag("m", "chars", wc.WcChars),
		boolFlag("L", "max-line-length", wc.WcMaxLineLength),
	}

	xargsFlags = flagSet{valueFlag("n", "max-args", intMaker(func(n int) any { return xargs.XargsMaxArgs(n) }))}

	grepFlags = flagSet{
		boolFlag("i", "ignore-case", grep.GrepIgnoreCase),
		boolFlag("v", "invert-match", grep.GrepInvert),
		boolFlag("x", "line-regexp", grep.GrepWholeLine),
		boolFlag("E", "extended-regexp", grep.GrepExtended),
		boolFlag("w", "word-regexp", grep.GrepWord),
		boolFlag("n", "line-number", grep.GrepLineNumbers),
		boolFlag("c", "count", grep.GrepCount),
	}

	trFlags = flagSet{
		boolFlag("d", "delete", tr.TrDelete),
		boolFlag("s", "squeeze-repeats", tr.TrSqueeze),
		boolFlag("c", "complement", tr.TrComplement),
	}

	seqFlags = flagSet{
		boolFlag("w", "equal-width", seq.SeqEqualWidth),
		valueFlag("s", "separator", strMaker(func(s string) any { return seq.SeqSeparator(s) })),
		valueFlag("f", "format", strMaker(func(s string) any { return seq.SeqFormat(s) })),
	}

	lsFlags = flagSet{
		boolFlag("a", "all", ls.LsAll),
		boolFlag("R", "recursive", ls.LsRecursive),
		boolFlag("l", "long", ls.LsLongFormat),
	}

	findFlags = flagSet{
		valueFlag("", "name", strMaker(func(s string) any { return findName(s) })),
		valueFlag("", "type", strMaker(func(s string) any { return findType(s) })),
		valueFlag("", "maxdepth", intMaker(func(n int) any { return findDepth(n) })),
	}

	perlFlags = flagSet{
		boolFlag("n", "loop", perl.PerlLoop),
		boolFlag("p", "loop-print", perl.PerlPrint),
		boolFlag("a", "autosplit", perl.PerlAutoSplit),
	}
)

// --- named builders for commands with positional or value-dependent shapes ---

// buildEcho builds the echo source, joining arguments with spaces.
func buildEcho(_ environment, positional Argv, _ []any) (segment, error) {
	return segment{source: echo.Echo(positional.strings()...)}, nil
}

// buildEmit builds the emit source from joined literal text.
func buildEmit(_ environment, positional Argv, _ []any) (segment, error) {
	return segment{source: emit.Emit(emit.EmitStdout(strings.Join(positional.strings(), " ")))}, nil
}

// buildYes builds the yes source, optionally repeating a custom line.
func buildYes(_ environment, positional Argv, _ []any) (segment, error) {
	if len(positional) == 0 {
		return segment{source: yes.Yes()}, nil
	}
	return segment{source: yes.Yes(yes.YesText(strings.Join(positional.strings(), " ")))}, nil
}

// buildSeq builds the seq source from numeric positionals plus flag options.
func buildSeq(_ environment, positional Argv, opts []any) (segment, error) {
	args := make([]any, 0, len(positional)+len(opts))
	for _, p := range positional {
		n, err := numArg(p)
		if err != nil {
			return segment{}, err
		}
		args = append(args, n)
	}
	args = append(args, opts...)
	return segment{source: seq.Seq(args...)}, nil
}

// buildLs builds the ls source. With no argument, or a single directory
// argument, it lists that directory's entries (the cmd-ls behavior). With
// multiple arguments or a single non-directory — typically the result of a glob
// like "*.go" — it emits the named paths themselves, matching how a shell's ls
// reports matched files.
func buildLs(env environment, positional Argv, opts []any) (segment, error) {
	if lsListsNames(env, positional) {
		return segment{source: gloo.SliceSource(toLines(positional))}, nil
	}
	opts = append(opts, ls.LsFs{Fs: env.fs})
	return segment{source: ls.Ls(firstOr(positional, "."), opts...)}, nil
}

// lsListsNames reports whether ls should echo its arguments as names rather than
// list a directory: true for multiple arguments or a single non-directory.
func lsListsNames(env environment, positional Argv) bool {
	if len(positional) == 0 {
		return false
	}
	if len(positional) > 1 {
		return true
	}
	isDir, _ := afero.DirExists(env.fs, string(positional[0]))
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
func buildFind(env environment, positional Argv, opts []any) (segment, error) {
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
	src := find.Find(firstOr(positional, "."),
		find.FindFs(env.fs),
		find.FindName(name),
		find.FindType(typ),
		find.FindMaxDepth(depth))
	return segment{source: src}, nil
}

// buildGrep builds the grep filter; the first positional is the pattern and the
// remainder are input files.
func buildGrep(_ environment, positional Argv, opts []any) (segment, error) {
	if len(positional) == 0 {
		return segment{}, ErrMissingArgument.With(nil, "grep: pattern")
	}
	cmd := grep.Grep(string(positional[0]), opts...)
	return segment{command: cmd, files: toFiles(positional[1:])}, nil
}

// buildSed builds the sed filter; the first positional is the expression and the
// remainder are input files.
func buildSed(_ environment, positional Argv, _ []any) (segment, error) {
	if len(positional) == 0 {
		return segment{}, ErrMissingArgument.With(nil, "sed: expression")
	}
	return segment{command: sed.Sed(string(positional[0])), files: toFiles(positional[1:])}, nil
}

// buildTr builds the tr filter. SET1 is required; SET2 is consumed only when the
// command is not in delete mode. Any trailing positionals are input files.
func buildTr(_ environment, positional Argv, opts []any) (segment, error) {
	if len(positional) == 0 {
		return segment{}, ErrMissingArgument.With(nil, "tr: SET1")
	}
	from := string(positional[0])
	rest := positional[1:]
	to := ""
	if trWantsSet2(opts) && len(rest) > 0 {
		to = string(rest[0])
		rest = rest[1:]
	}
	return segment{command: tr.Tr(from, to, opts...), files: toFiles(rest)}, nil
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
func buildShuf(_ environment, positional Argv, opts []any) (segment, error) {
	clean, echo := extractEcho(opts)
	if echo {
		clean = append(clean, shuf.ShufEcho(positional.strings()...))
		return segment{command: shuf.Shuf(clean...)}, nil
	}
	return segment{command: shuf.Shuf(clean...), files: toFiles(positional)}, nil
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
func nlBody(value Argument) (any, error) {
	switch value {
	case "a":
		return nl.NlBodyAll, nil
	case "t":
		return nl.NlBodyNonEmpty, nil
	case "n":
		return nl.NlBodyNone, nil
	}
	return nil, ErrInvalidFlagValue.With(nil, "nl -b: "+string(value))
}
