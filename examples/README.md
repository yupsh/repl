# yupsh examples

A full suite of runnable yupsh sessions, one file per theme, plus a single runner that executes them all.

Each `*.yup` file is a real yupsh session: comment lines (`#`) and blank lines are ignored by the REPL, and every command line is fed to the binary on stdin. Pipe one in directly:

```sh
go build -o yupsh ./cmd/yupsh
./yupsh < examples/01-sources.yup
```

Or run the whole suite with the runner (builds the binary, seeds a deterministic fixture tree, runs every file, and fails if any example writes to stderr):

```sh
examples/run.sh          # or: make examples
YUPSH=./bin/yupsh examples/run.sh   # reuse an existing binary instead of building
```

## The suite

| File | Shows |
| --- | --- |
| [`01-sources.yup`](01-sources.yup) | `echo`, `emit`, `seq` (flags), `yes` bounded by `head` |
| [`02-pipelines.yup`](02-pipelines.yup) | multi-stage `\|` composition, `tee`, `tr`, extended-regexp `grep` |
| [`03-text.yup`](03-text.yup) | `rev`, `tr`, `sed`, `nl`, `cat -n`, `sort`, `uniq -c` |
| [`04-encoding.yup`](04-encoding.yup) | `base64` round-trip, `hexdump`, `json` |
| [`05-fields-and-sets.yup`](05-fields-and-sets.yup) | `cut`, `paste`, `split`, `xargs`, `comm`, `diff`, `join` |
| [`06-paths.yup`](06-paths.yup) | `basename`, `dirname` |
| [`07-search-and-sample.yup`](07-search-and-sample.yup) | `grep` flags, `head`/`tail` `-NUM`, seeded `shuf` |
| [`08-files-and-globs.yup`](08-files-and-globs.yup) | `ls`, `find`, globbing, `~` expansion, file input |
| [`09-external.yup`](09-external.yup) | `exec`, `git` |
| [`10-perl.yup`](10-perl.yup) | `perl` one-liners (the runner skips this if `perl` is absent) |

## Why every line starts with a source

In a piped (non-interactive) session the REPL's line reader and a command's stdin are the **same** stream. A pipeline that starts with a bare filter would read the *rest of this file* as its input, so every example begins with a source (`echo`/`emit`/`seq`/`ls`/`find`/`yes`) or reads a named file. Interactively (at a TTY) the two streams are distinct, so any command works on its own.

## Future: run the suite with yupsh's own `xargs`

The goal is for this runner to become a yupsh-native pipeline that uses yupsh's **own** `xargs` to run every example. [`cmd-xargs`](https://github.com/gloo-foo/cmd-xargs) now has an exec mode — given a command, `xargs` runs it once per argument group, resolving the command against the registry (and falling back to a subprocess for unknown commands) — so the per-example loop in [`run.sh`](run.sh) can be replaced by a yupsh pipeline that feeds each example path through `xargs` to the runner. [`run.sh`](run.sh) still drives the binary from bash until that migration lands.
