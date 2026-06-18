// Package line orchestrates execution of a single REPL input line.
//
// It defines the command's Config (the injected filesystem and home directory)
// and Run (the orchestration entry point the app tier invokes). Run tokenizes
// the line, expands each segment's arguments, resolves and builds each segment
// against the command registry, and plans the assembled pipeline — delegating
// every step to the reusable internal/token, internal/expansion, internal/flags,
// internal/command, and internal/pipeline packages. It contains no tokenizing,
// flag-parsing, or stream-running logic of its own. This is the domain tier: the
// seam between the app tier (internal/app) and the implementation packages.
//
// Run returns the planned internal/pipeline.Assembly rather than running it: the
// app tier "renders" that result by streaming it to the session's output writer,
// mirroring how a gomatic/template.cli domain returns a Result the app renders.
package line
