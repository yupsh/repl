# Development

## Prerequisites

- Go 1.26+
- The sibling `gloo-foo/framework` and `gloo-foo/cmd-*` checkouts (wired via
  `replace` directives in `go.mod` during local development).

## Build & run

```bash
cd repl
go build -o yupsh ./cmd/yupsh
./yupsh
```

## Quality gate

The module pins its tooling in the `go.mod` `tool` stanza and gates every change
through `make check` — it must exit zero:

```bash
make check    # gofumpt, go vet, staticcheck, gocognit (<=7), govulncheck, 100% coverage
```

Every `internal/...` package is held to **100% statement coverage**. The
`cmd/yupsh` entry point is the single os-wiring seam, excluded from coverage like
the framework's own gate.

## Integration tests

Black-box tests build the real `yupsh` binary and drive it through stdin,
asserting the tool's user-facing claims — pipelines compose, flags are
translated, globs and `~` expand, files and quoting work, built-ins and errors
behave. They deliberately do **not** re-test the underlying `cmd-*` behavior
(that `wc` counts, that `grep` matches) — each command owns that.

They are gated behind the `integration` build tag so they are intentional and
excluded from the hermetic `make check`:

```bash
make integration          # or: go test -tags integration ./...
```

The aim is high confidence that the tool does what it claims, and that future
changes don't silently break a claim — if a claim changes, its test changes too.

## Contributing

1. Branch from `main`.
2. Make the change; keep every `internal/...` package at 100% coverage and within the gate.
3. If you add or change a user-facing capability, add or update its
   black-box case in `integration_test.go`.
4. Run `make check` **and** `make integration` — both green.
5. Open a PR with a clear description.

When reporting an issue, include the `version` output, your Go version, OS, and
steps to reproduce.
