# yupsh — Yup Shell

The interactive REPL for the [yup.sh](https://github.com/yupsh) /
[gloo-foo](https://github.com/gloo-foo) command ecosystem: type-safe Go commands
composed into Unix-style pipelines, with globbing, `~` expansion, history, and
in-line editing. The program is `yupsh` (Yup **Sh**ell); the prompt is `yup>`.

```bash
go install github.com/yupsh/repl/yupsh@latest
yupsh
```

Full documentation — usage, the command catalogue, architecture, and
development — is maintained in the project **wiki**, kept separate so improving
the docs does not churn the codebase.

## Develop

```bash
make help         # list every target
make check        # gofumpt, vet, staticcheck, gocognit (<=7), 100% coverage, goreleaser check, govulncheck
make integration  # black-box integration tests (go:build integration)
make build        # build the yupsh binary for the current platform (via goreleaser)
```

Releases are cut by [goreleaser](https://goreleaser.com): pushing a `v*` tag runs
the [release workflow](.github/workflows/release.yml), which builds the
linux/darwin × amd64/arm64 archives described in [.goreleaser.yaml](.goreleaser.yaml).

## License

See [LICENSE](LICENSE).
