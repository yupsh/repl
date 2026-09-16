# Architecture

yupsh is laid out to the [gomatic/template.cli](https://github.com/gomatic/template.cli) three-tier standard — **app → orchestration → implementation** (template.cli's domain tier, here the line orchestrator) — adapted to an interactive REPL. Dependencies flow one direction only: a tier depends only on the tier to its right, and everything below the entry point lives under `internal/` (the binary is the only importable artifact).

```text
repl/
├── cmd/yupsh/main.go        — entry point: wires os stdin/stdout/stderr + afero.NewOsFs
│                              into the app.Session; on a TTY drives it through a
│                              golang.org/x/term.Terminal for history/editing, else
│                              the plain scanner path. The single os-wiring seam.
├── internal/
│   ├── app/                 — APP TIER: the REPL surface — the read-eval-print loop,
│   │                          built-ins (exit/help/version/clear), banner, help, and
│   │                          the LineReader. Hands each line to the line orchestrator.
│   ├── line/                — ORCHESTRATION TIER: orchestrates one input line into a
│   │                          planned pipeline.Assembly (parse → expand → resolve →
│   │                          build → plan), delegating to implementation packages.
│   ├── token/               — IMPLEMENTATION: tokenize a line into pipeline segments
│   ├── expansion/           — IMPLEMENTATION: tilde + glob expansion (POSIX-style)
│   ├── flags/               — IMPLEMENTATION: Unix flag → typed-option translation
│   ├── command/             — IMPLEMENTATION: the command registry + segment builders
│   ├── pipeline/            — IMPLEMENTATION: assemble + run the typed-stream pipeline
│   └── constants/           — sentinel Error type and the Err* constants
├── integration_test.go      — black-box tests of the compiled binary (build-tagged)
├── go.mod                   — module dependencies and tool gate
└── Makefile                 — quality gate (make check) + integration target
```

## The three tiers

| Tier | Location | Responsibility | Forbidden |
| --- | --- | --- | --- |
| **app** | [`internal/app`](../internal/app) | The REPL surface: the loop, built-ins, banner, help, line reading, and rendering a planned pipeline to the output writer. | Tokenizing, expansion, flag translation, pipeline assembly. |
| **orchestration** | [`internal/line`](../internal/line) | Orchestration: a `Config` (injected `Fs` + `Home`) and a thin `Run` that turns a line into a `pipeline.Assembly` by composing the implementation packages. | Importing the app tier; output formatting; reusable work that belongs in an implementation package. |
| **implementation** | [`internal/<concept>`](../internal) | The actual, reusable work, named for the concept (`token`, `expansion`, `flags`, `command`, `pipeline`). | Any knowledge of the REPL or of being "a command". |

`internal/app` is a reserved name, and [`internal/line`](../internal/line) is the one orchestration package (the REPL has a single "command": the input line — there is no `internal/app/commands` tree, so the orchestrator sits beside the implementation packages rather than under a `domain/` prefix). Every other `internal/<name>` is an implementation package named for the concept it implements. Errors are constant sentinels in [`internal/constants`](../internal/constants), matched with `errors.Is`.

### Why a hand-written parser, not urfave/cli

template.cli targets `urfave/cli` apps, which parse `os.Args` once against a fixed command tree. yupsh is the opposite: it interprets _live shell lines_ — pipes, quoting, runtime globbing, and per-command flag translation onto third-party `cmd-*` constructors — none of which `urfave/cli` models. So yupsh keeps its own tokenizer/expansion/flag layer and maps only the template's _structure_ (the three tiers, `cmd/` entry, `internal/constants`, the gate), not its argument parser.

### The app↔orchestrator seam

A template.cli domain `Run` returns a serializable `Result` that the app tier renders as JSON. yupsh streams bytes instead, so the seam differs by design: [`line.Run`](../internal/line/run.go) returns the planned [`pipeline.Assembly`](../internal/pipeline/pipeline.go) (the "result"), and [`app.Session.execute`](../internal/app/session.go) "renders" it by calling `Assembly.Run(ctx, out)` to stream the pipeline. The orchestrator holds no output writer; the app holds no business logic.

## How a line is executed

1. [`token.Parse`](../internal/token/token.go) tokenizes the line, recording quoting, and splits on `|`.
2. [`expansion.Expand`](../internal/expansion/expansion.go) applies tilde then glob expansion to each command's arguments (unquoted tokens only), against the injected `afero.Fs`.
3. [`flags.Parse`](../internal/flags/flags.go) translates each segment's flags against the command's flag table into the typed options its `cmd-*` constructor expects.
4. [`command`](../internal/command) builds each segment; the first stage selects the input source (source command, files, or stdin), later stages become filters.
5. [`pipeline.Plan`](../internal/pipeline/pipeline.go) folds the stages into an `Assembly`, which `app` runs via `gloo.RunContext(ctx, source, gloo.ByteWriteTo(out), …)`.

## Adding a command

Add one entry to the map in [`internal/command/registry.go`](../internal/command/registry.go). Most filters are a single line:

```go
"wc": {Flags: wcFlags, Build: filter(func(o []any) gloo.Command[[]byte, []byte] {
    return wc.Wc(o...)
}), Summary: "count lines, words, and bytes"},
```

Declare its flags as a `flags.Set` of `flags.Bool`/`flags.Value`/`flags.Num` entries that map each Unix flag to the typed option, cover the new builder and makers in [`command_test.go`](../internal/command/command_test.go) to 100%, and add a behavioral case to the [integration tests](Development) if it introduces a new capability.

During local development the `gloo-foo/framework` and `gloo-foo/cmd-*` modules are wired via the gitignored `go.work` to their sibling checkouts.

## Quality gate

`make check` ([`Makefile`](../Makefile)) must exit zero before any change is complete: `gofumpt`, `go vet`, `staticcheck`, `gocognit -over 7`, **100% statement coverage** of every `internal/...` package, `goreleaser check`, and `govulncheck`. The `cmd/yupsh` entry point is the single os-wiring seam, excluded from coverage like the framework's own gate.
