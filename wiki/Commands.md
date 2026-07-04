# Commands

Run `help` in the shell for the live catalogue. Sources are marked `(source)`.

- **Generate**: `echo`, `emit`, `seq`, `yes` *(sources)*
- **Filesystem**: `ls`, `find` *(sources)*, `basename`, `dirname`
- **Search & filter**: `grep`, `comm`, `uniq`
- **Transform**: `tr`, `sed`, `perl`, `rev`, `cut`, `paste`, `split`, `join`,
  `nl`, `cat`, `tac`, `tail`, `head`
- **Summarize**: `wc`, `sort`, `shuf`, `diff`, `hexdump`, `base64`, `json`
- **External**: `exec`, `git`

## Flags

Flags are the responsibility of each `cmd-*` package; the REPL **translates** the
Unix-style flags you type into the typed options those packages expose. For
example `wc -l` → `wc.WcLines`, `head -n 10` / `head -10` → `head.HeadLines(10)`,
`grep -v` → `grep.GrepInvert`. See each command's own repository for the full
flag semantics.

## Not wired

Some constructors require non-textual Go values and therefore cannot be built
from a typed command line:

- `awk` — takes a `Program` **interface** (Begin/Condition/Action/End); there is
  no string→`Program` parser.
- `while` — takes a `func([]byte) ([]byte, error)`.
- `capture` — takes `io.Writer` sinks.

If `cmd-awk` later offers a parser that builds a `Program` from a script string,
`awk` can be wired like the others.
