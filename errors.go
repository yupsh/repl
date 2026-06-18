package repl

import "fmt"

// Error is the single sentinel error type for the repl package. Every error the
// package emits is a constant of this type, so callers test with errors.Is
// rather than comparing strings.
type Error string

// Error satisfies the error interface.
func (e Error) Error() string { return string(e) }

// With wraps an optional cause and optional context arguments around the
// sentinel while keeping errors.Is(result, e) true. It mirrors the framework's
// gloo.Error.With so error construction stays uniform across the ecosystem.
func (e Error) With(cause error, args ...any) error {
	out := error(e)
	if cause != nil {
		out = fmt.Errorf("%w: %w", e, cause)
	}
	if len(args) > 0 {
		out = fmt.Errorf("%w: %s", out, fmt.Sprint(args...))
	}
	return out
}

const (
	// ErrEmptyCommand is returned for a pipeline segment with no command name.
	ErrEmptyCommand Error = "empty command"
	// ErrUnknownCommand is returned for a command name not in the registry.
	ErrUnknownCommand Error = "unknown command"
	// ErrUnterminatedQuote is returned when a line ends inside a quoted span.
	ErrUnterminatedQuote Error = "unterminated quote"
	// ErrSourceMidPipeline is returned when a source command (echo, seq, ls,
	// find, yes, emit) appears after a pipe, where only filters are valid.
	ErrSourceMidPipeline Error = "source command cannot appear after a pipe"
	// ErrArgsMidPipeline is returned when a non-first segment carries positional
	// arguments (files or literal input); only the first segment may source
	// input into the pipeline.
	ErrArgsMidPipeline Error = "positional arguments are only allowed on the first command"
	// ErrMissingArgument is returned when a required positional is absent.
	ErrMissingArgument Error = "missing required argument"
	// ErrUnknownFlag is returned for a flag not declared by the command.
	ErrUnknownFlag Error = "unknown flag"
	// ErrFlagNeedsValue is returned when a value-taking flag has no value.
	ErrFlagNeedsValue Error = "flag requires a value"
	// ErrInvalidNumber is returned when a numeric argument fails to parse.
	ErrInvalidNumber Error = "invalid number"
	// ErrInvalidFlagValue is returned when a flag value is malformed.
	ErrInvalidFlagValue Error = "invalid flag value"
)
