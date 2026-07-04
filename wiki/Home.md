# Yup Shell (yupsh)

An interactive shell — a REPL — for the [gloo-foo](https://github.com/gloo-foo)
ecosystem. It composes the [gloo-foo/framework](https://github.com/gloo-foo/framework)
typed-stream pipeline out of the [`gloo-foo/cmd-*`](https://github.com/gloo-foo)
command modules, giving a familiar shell interface backed entirely by type-safe
Go commands — no external processes for the core tools.

The prompt is `yup>`; the program is `yupsh` (Yup **Sh**ell, like `bash`/`fish`).

## Pages

- **[Usage](Usage)** — running the shell, the input model, examples
- **[Commands](Commands)** — the command catalogue and what is/isn't wired
- **[Architecture](Architecture)** — how a line becomes a pipeline; adding a command
- **[Development](Development)** — the quality gate, integration tests, contributing

## Features

- **Typed-stream pipelines** — every stage is a `gloo.Command[[]byte, []byte]`,
  composed with `|` and run through the framework's `RunContext`.
- **Source and filter commands** — sources (`echo`, `seq`, `ls`, `find`, `yes`,
  `emit`) start a pipeline; filters (`grep`, `wc`, `sort`, …) transform it.
- **Unix-style flags** — `wc -l`, `grep -v`, `head -n 10`, `cut -d , -f 2`, the
  GNU `head -5`/`tail -3` shorthand — translated to each command's typed options.
- **Shell expansion** — globbing (`*`, `?`, `[…]`) and `~` expand against the
  filesystem; quotes suppress expansion; non-matching globs stay literal.
- **File or stdin input** — file arguments on the first command are opened as the
  pipeline's input; otherwise input is read from stdin.
- **Interactive history & editing** — on a TTY, up/down history and in-line
  editing via `golang.org/x/term`; piped input uses plain scanning.
- **Built-ins** — `help`, `version`, `clear`, `exit`/`quit`, and `#` comments.

## Install

```bash
go install github.com/yupsh/repl/yupsh@latest
```

Installs a `yupsh` binary to `$GOPATH/bin` (or `$HOME/go/bin`).

## Quick start

```text
yupsh REPL v0.2.0
Yup Shell — a typed-stream REPL over gloo-foo/framework and cmd-* commands.
Type 'help' for commands, 'exit' to quit.

yup> echo hello world
hello world
yup> seq 1 10 | grep -v 5 | head -n 3
1
2
3
yup> wc -l *.go
```
