# Architecture

```text
repl/
├── *.go            — the repl package (library): tokenizer, expansion (glob +
│                     tilde), flag translation, per-command registry, pipeline
│                     assembly, the REPL engine, line reader
├── yupsh/main.go   — thin main: wires os I/O + afero.NewOsFs into the engine;
│                     on a TTY drives it through a golang.org/x/term.Terminal
│                     for history/editing, else the scanner path
├── integration_test.go — black-box tests of the compiled binary (build-tagged)
├── go.mod          — module dependencies and tool gate
└── Makefile        — quality gate (make check) + integration target
```

All behaviour lives in the `repl` package and is covered to 100%. `main` is the
single os-wiring seam.

## How a line is executed

1. `parseLine` tokenizes the line, recording quoting, and splits on `|`.
2. `expandArgs` applies tilde then glob expansion to each command's arguments
   (unquoted tokens only), against the injected `afero.Fs`.
3. Each segment's flags are translated against the command's flag table to the
   typed options its `cmd-*` constructor expects.
4. The first stage selects the input source (source command, files, or stdin);
   later stages become filters.
5. The pipeline runs via `gloo.RunContext(ctx, source, gloo.ByteWriteTo(out), …)`.

## Adding a command

Add one entry to the map in `registry.go`. Most filters are a single line:

```go
"wc": {flags: wcFlags, build: filter(func(o []any) gloo.Command[[]byte, []byte] {
    return wc.Wc(o...)
}), summary: "count lines, words, and bytes"},
```

Declare its flags as a `flagSet` of `boolFlag`/`valueFlag`/`numFlag` entries that
map each Unix flag to the typed option, then add a behavioral case to the
[integration tests](Development) if it introduces a new capability.

During local development the `gloo-foo/framework` and `gloo-foo/cmd-*` modules
are wired via `replace` directives to their sibling checkouts (see `go.mod`).
