# Contributing

Development setup, the quality gate, the integration tests, and the contribution
workflow are documented in the project **wiki**.

In short: keep the `repl` package at 100% coverage, run `make check` and
`make integration` (both green), and add a black-box case in
`integration_test.go` for any new user-facing capability.
