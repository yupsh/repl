// Package constants declares every sentinel error the yupsh REPL can emit. The
// sentinel mechanism is gomatic/go-error's errs.Const; this package declares
// only the values, so callers test with errors.Is rather than comparing strings.
package constants

import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	// ErrArgsMidPipeline is returned when a non-first segment carries positional
	// arguments (files or literal input); only the first segment may source
	// input into the pipeline.
	ErrArgsMidPipeline errs.Const = "positional arguments are only allowed on the first command"
	// ErrEmptyCommand is returned for a pipeline segment with no command name.
	ErrEmptyCommand errs.Const = "empty command"
	// ErrFlagNeedsValue is returned when a value-taking flag has no value.
	ErrFlagNeedsValue errs.Const = "flag requires a value"
	// ErrInvalidFlagValue is returned when a flag value is malformed.
	ErrInvalidFlagValue errs.Const = "invalid flag value"
	// ErrInvalidNumber is returned when a numeric argument fails to parse.
	ErrInvalidNumber errs.Const = "invalid number"
	// ErrMissingArgument is returned when a required positional is absent.
	ErrMissingArgument errs.Const = "missing required argument"
	// ErrSourceMidPipeline is returned when a source command (echo, seq, ls,
	// find, yes, emit) appears after a pipe, where only filters are valid.
	ErrSourceMidPipeline errs.Const = "source command cannot appear after a pipe"
	// ErrUnknownCommand is returned for a command name not in the registry.
	ErrUnknownCommand errs.Const = "unknown command"
	// ErrUnknownFlag is returned for a flag not declared by the command.
	ErrUnknownFlag errs.Const = "unknown flag"
	// ErrUnterminatedQuote is returned when a line ends inside a quoted span.
	ErrUnterminatedQuote errs.Const = "unterminated quote"
)
