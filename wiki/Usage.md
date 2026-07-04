# Usage

Start the shell:

```bash
yupsh
```

Or drive it non-interactively (pipes and redirects use plain line scanning, so
scripts are deterministic):

```bash
yupsh < script.txt
```

## Input model

A pipeline's input is chosen by its **first** stage:

| First stage | Input source |
|-------------|--------------|
| A source command (`echo`, `seq`, `ls`, `find`, `yes`, `emit`) | the command generates the stream |
| A filter with file arguments (`cat file.txt`, `wc -l a.txt`) | the named files, opened via the framework's `ByteFileSource` |
| A filter with no file arguments (`grep foo`) | standard input |

Later stages must be filters — a source after a pipe, or file arguments on a
non-first stage, are reported as errors.

## Shell expansion

- **Globbing**: unquoted `*`, `?`, and `[…]` expand against the working
  directory. A pattern with no matches is left literal (POSIX default).
- **Tilde**: a leading `~` or `~/…` expands to the home directory.
- **Quoting**: single or double quotes group arguments and suppress expansion
  (`echo '*.go'` prints `*.go`).

```text
yup> ls *.go
yup> wc -l *.md
yup> cat ~/notes.txt | grep TODO
```

## Examples

```text
# sources
yup> echo hello world
yup> seq 1 10

# pipelines
yup> seq 1 10 | grep -v 5 | head -n 3
yup> seq 1 100 | wc -l
yup> echo HELLO | tr A-Z a-z

# files and globs
yup> wc -l *.go
yup> cat data.txt | sort | uniq

# subprocess escape hatches
yup> echo hi | exec cat
yup> echo hello | perl -p 's/l/L/g'
```

## Built-ins

- `help` — list commands and usage
- `version` — show version
- `clear` — clear the screen
- `exit` / `quit` — leave the shell
- lines beginning with `#` are ignored
