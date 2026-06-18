// Package repl is the module root marker for the yupsh REPL.
//
// The program is laid out to the gomatic/template.cli three-tier standard: the
// binary entry point is cmd/yupsh, and all behaviour lives under internal/ —
// the app tier (internal/app), the line domain (internal/domain/line), and the
// implementation packages (internal/token, internal/expansion, internal/flags,
// internal/command, internal/pipeline) over the shared internal/constants
// sentinel errors. See docs/architecture.md.
package repl
