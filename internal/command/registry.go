package command

import (
	gloo "github.com/gloo-foo/framework"

	base64 "github.com/gloo-foo/cmd-base64"
	basename "github.com/gloo-foo/cmd-basename"
	cat "github.com/gloo-foo/cmd-cat"
	comm "github.com/gloo-foo/cmd-comm"
	cut "github.com/gloo-foo/cmd-cut"
	diff "github.com/gloo-foo/cmd-diff"
	dirname "github.com/gloo-foo/cmd-dirname"
	exec "github.com/gloo-foo/cmd-exec"
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

	"github.com/yupsh/repl/internal/flags"
)

// Registry returns the command table. It is built fresh per caller so callers
// never share mutable state.
func Registry() map[Name]Builder {
	return map[Name]Builder{
		"echo":     {Raw: true, Build: buildEcho, Summary: "emit arguments as a line (source)"},
		"emit":     {Raw: true, Build: buildEmit, Summary: "emit literal text (source)"},
		"yes":      {Raw: true, Build: buildYes, Summary: "repeat a line until interrupted (source)"},
		"seq":      {Flags: seqFlags, Build: buildSeq, Summary: "generate a numeric sequence (source)"},
		"ls":       {Flags: lsFlags, Build: buildLs, Summary: "list directory entries (source)"},
		"find":     {Flags: findFlags, Build: buildFind, Summary: "walk a directory tree (source)"},
		"base64":   {Flags: base64Flags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return base64.Base64(o...) }), Summary: "base64 encode or decode"},
		"basename": {Flags: basenameFlags, Build: literal(func(o []any) gloo.Command[[]byte, []byte] { return basename.Basename(o...) }), Summary: "strip directory and suffix"},
		"cat":      {Flags: catFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return cat.Cat(o...) }), Summary: "concatenate with optional numbering"},
		"comm":     {Flags: commFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return comm.Comm(o...) }), Summary: "compare sorted streams"},
		"cut":      {Flags: cutFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return cut.Cut(o...) }), Summary: "select fields, bytes, or characters"},
		"diff":     {Flags: diffFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return diff.Diff(o...) }), Summary: "compare streams line by line"},
		"dirname":  {Build: literal(func(o []any) gloo.Command[[]byte, []byte] { return dirname.Dirname(o...) }), Summary: "strip last path component"},
		"head":     {Flags: headFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return head.Head(o...) }), Summary: "output the first lines or bytes"},
		"hexdump":  {Flags: hexdumpFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return hexdump.Hexdump(o...) }), Summary: "hex dump of input"},
		"join":     {Flags: joinFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return join.Join(o...) }), Summary: "join lines on a common field"},
		"json":     {Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return jsoncmd.Json(o...) }), Summary: "pretty-print JSON"},
		"nl":       {Flags: nlFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return nl.Nl(o...) }), Summary: "number lines"},
		"paste":    {Flags: pasteFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return paste.Paste(o...) }), Summary: "merge lines"},
		"rev":      {Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return rev.Rev(o...) }), Summary: "reverse each line"},
		"shuf":     {Flags: shufFlags, Build: buildShuf, Summary: "shuffle lines randomly"},
		"sort":     {Flags: sortFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return sortcmd.Sort(o...) }), Summary: "sort lines"},
		"split":    {Flags: splitFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return split.Split(o...) }), Summary: "split fields onto lines"},
		"tac":      {Flags: tacFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tac.Tac(o...) }), Summary: "reverse line order"},
		"tail":     {Flags: tailFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tail.Tail(o...) }), Summary: "output the last lines or bytes"},
		"tee":      {Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return tee.Tee(o...) }), Summary: "pass input through unchanged"},
		"uniq":     {Flags: uniqFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return uniq.Uniq(o...) }), Summary: "drop adjacent duplicate lines"},
		"wc":       {Flags: wcFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] { return wc.Wc(o...) }), Summary: "count lines, words, and bytes"},
		"xargs":    {Flags: xargsFlags, Build: buildXargs, Summary: "run a command per argument group (or group fields into lines)"},
		"grep":     {Flags: grepFlags, Build: buildGrep, Summary: "filter lines matching a pattern"},
		"sed":      {Build: buildSed, Summary: "apply an s/// substitution"},
		"tr":       {Flags: trFlags, Build: buildTr, Summary: "translate, delete, or squeeze characters"},
		"exec":     {Raw: true, Build: command(func(p flags.Argv, _ []any) gloo.Command[[]byte, []byte] { return exec.Exec(p.Anys()...) }), Summary: "run an external program"},
		"git":      {Raw: true, Build: command(func(p flags.Argv, _ []any) gloo.Command[[]byte, []byte] { return git.Git(p.Strings()...) }), Summary: "run git"},
		"perl":     {Flags: perlFlags, Build: command(func(p flags.Argv, o []any) gloo.Command[[]byte, []byte] { return perl.Perl(append(p.Anys(), o...)...) }), Summary: "run a perl one-liner"},
	}
}

// --- flag tables ---

var (
	base64Flags = flags.Set{flags.Bool("d", "decode", base64.Base64Decode)}

	basenameFlags = flags.Set{flags.Value("s", "suffix", flags.StrMaker(func(s string) any { return basename.BasenameSuffix(s) }))}

	catFlags = flags.Set{
		flags.Bool("n", "number", cat.CatNumberLines),
		flags.Bool("b", "number-nonblank", cat.CatNumberNonBlank),
	}

	commFlags = flags.Set{
		flags.Bool("1", "", comm.CommSuppressColumn1),
		flags.Bool("2", "", comm.CommSuppressColumn2),
		flags.Bool("3", "", comm.CommSuppressColumn3),
	}

	cutFlags = flags.Set{
		flags.Value("d", "delimiter", flags.StrMaker(func(s string) any { return cut.CutDelimiter(s) })),
		flags.Value("f", "fields", func(v flags.Argument) (any, error) {
			xs, err := flags.IntList(v)
			if err != nil {
				return nil, err
			}
			return cut.CutFields(xs...), nil
		}),
		flags.Value("b", "bytes", flags.StrMaker(func(s string) any { return cut.CutBytes(s) })),
		flags.Value("c", "characters", flags.StrMaker(func(s string) any { return cut.CutChars(s) })),
		flags.Bool("", "complement", cut.CutComplement),
	}

	diffFlags = flags.Set{flags.Bool("u", "unified", diff.DiffUnified)}

	headFlags = flags.Set{
		flags.Value("n", "lines", flags.IntMaker(func(n int) any { return head.HeadLines(n) })),
		flags.Value("c", "bytes", flags.IntMaker(func(n int) any { return head.HeadBytes(n) })),
		flags.Num(flags.IntMaker(func(n int) any { return head.HeadLines(n) })),
	}

	hexdumpFlags = flags.Set{flags.Bool("C", "canonical", hexdump.HexdumpCanonical)}

	joinFlags = flags.Set{flags.Value("t", "separator", flags.StrMaker(func(s string) any { return join.JoinSeparator(s) }))}

	nlFlags = flags.Set{
		flags.Value("b", "body-numbering", nlBody),
		flags.Value("s", "separator", flags.StrMaker(func(s string) any { return nl.NlSep(s) })),
		flags.Value("v", "starting-line", flags.IntMaker(func(n int) any { return nl.NlStart(n) })),
		flags.Value("i", "increment", flags.IntMaker(func(n int) any { return nl.NlIncrement(n) })),
		flags.Value("w", "width", flags.IntMaker(func(n int) any { return nl.NlWidth(n) })),
		flags.Value("n", "number-format", flags.StrMaker(func(s string) any { return nl.NlFormat(s) })),
	}

	pasteFlags = flags.Set{
		flags.Value("d", "delimiters", flags.StrMaker(func(s string) any { return paste.PasteDelimiter(s) })),
		flags.Bool("s", "serial", paste.PasteSerial),
	}

	sortFlags = flags.Set{
		flags.Bool("r", "reverse", sortcmd.SortReverse),
		flags.Bool("n", "numeric-sort", sortcmd.SortNumeric),
		flags.Bool("u", "unique", sortcmd.SortUnique),
		flags.Bool("f", "ignore-case", sortcmd.SortIgnoreCase),
		flags.Bool("R", "random-sort", sortcmd.SortRandom),
		flags.Bool("b", "ignore-leading-blanks", sortcmd.SortIgnoreLeadingBlanks),
		flags.Bool("V", "version-sort", sortcmd.SortVersionSort),
		flags.Bool("h", "human-numeric-sort", sortcmd.SortHumanNumeric),
		flags.Bool("M", "month-sort", sortcmd.SortMonthSort),
		flags.Bool("s", "stable", sortcmd.SortStableSort),
		flags.Value("k", "key", flags.IntMaker(func(n int) any { return sortcmd.SortField(n) })),
		flags.Value("t", "field-separator", flags.StrMaker(func(s string) any { return sortcmd.SortDelimiter(s) })),
	}

	splitFlags = flags.Set{flags.Value("d", "delimiter", flags.StrMaker(func(s string) any { return split.SplitDelim(s) }))}

	tacFlags = flags.Set{flags.Value("s", "separator", flags.StrMaker(func(s string) any { return tac.TacSep(s) }))}

	tailFlags = flags.Set{
		flags.Value("n", "lines", flags.IntMaker(func(n int) any { return tail.TailLines(n) })),
		flags.Value("c", "bytes", flags.IntMaker(func(n int) any { return tail.TailBytes(n) })),
		flags.Num(flags.IntMaker(func(n int) any { return tail.TailLines(n) })),
	}

	uniqFlags = flags.Set{
		flags.Bool("d", "repeated", uniq.UniqDuplicatesOnly),
		flags.Bool("c", "count", uniq.UniqCount),
	}

	shufFlags = flags.Set{
		flags.Value("n", "head-count", flags.IntMaker(func(n int) any { return shuf.ShufCount(n) })),
		flags.Value("", "seed", flags.Int64Maker(func(n int64) any { return shuf.ShufSeed(n) })),
		flags.Value("i", "input-range", func(v flags.Argument) (any, error) {
			lo, hi, err := flags.RangeArg(v)
			if err != nil {
				return nil, err
			}
			return shuf.ShufRange(lo, hi), nil
		}),
		flags.Bool("e", "echo", shufEchoMarker{}),
	}

	wcFlags = flags.Set{
		flags.Bool("l", "lines", wc.WcLines),
		flags.Bool("w", "words", wc.WcWords),
		flags.Bool("c", "bytes", wc.WcBytes),
		flags.Bool("m", "chars", wc.WcChars),
		flags.Bool("L", "max-line-length", wc.WcMaxLineLength),
	}

	xargsFlags = flags.Set{
		flags.Value("n", "max-args", flags.IntMaker(func(n int) any { return xargs.XargsMaxArgs(n) })),
		flags.Value("I", "replace", flags.StrMaker(func(s string) any { return xargs.XargsReplace(s) })),
		flags.Value("P", "max-procs", flags.IntMaker(func(n int) any { return xargs.XargsMaxProcs(n) })),
		flags.Bool("0", "null", xargs.XargsNull(true)),
	}

	grepFlags = flags.Set{
		flags.Bool("i", "ignore-case", grep.GrepIgnoreCase),
		flags.Bool("v", "invert-match", grep.GrepInvert),
		flags.Bool("x", "line-regexp", grep.GrepWholeLine),
		flags.Bool("E", "extended-regexp", grep.GrepExtended),
		flags.Bool("w", "word-regexp", grep.GrepWord),
		flags.Bool("n", "line-number", grep.GrepLineNumbers),
		flags.Bool("c", "count", grep.GrepCount),
	}

	trFlags = flags.Set{
		flags.Bool("d", "delete", tr.TrDelete),
		flags.Bool("s", "squeeze-repeats", tr.TrSqueeze),
		flags.Bool("c", "complement", tr.TrComplement),
	}

	seqFlags = flags.Set{
		flags.Bool("w", "equal-width", seq.SeqEqualWidth),
		flags.Value("s", "separator", flags.StrMaker(func(s string) any { return seq.SeqSeparator(s) })),
		flags.Value("f", "format", flags.StrMaker(func(s string) any { return seq.SeqFormat(s) })),
	}

	lsFlags = flags.Set{
		flags.Bool("a", "all", ls.LsAll),
		flags.Bool("R", "recursive", ls.LsRecursive),
		flags.Bool("l", "long", ls.LsLongFormat),
	}

	findFlags = flags.Set{
		flags.Value("", "name", flags.StrMaker(func(s string) any { return findName(s) })),
		flags.Value("", "type", flags.StrMaker(func(s string) any { return findType(s) })),
		flags.Value("", "maxdepth", flags.IntMaker(func(n int) any { return findDepth(n) })),
	}

	perlFlags = flags.Set{
		flags.Bool("n", "loop", perl.PerlLoop),
		flags.Bool("p", "loop-print", perl.PerlPrint),
		flags.Bool("a", "autosplit", perl.PerlAutoSplit),
	}
)
