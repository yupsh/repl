package constants

// Keep these constants sorted alphabetically.
const (
	// ErrArgsMidPipeline is returned when a non-first segment carries positional
	// arguments (files or literal input); only the first segment may source
	// input into the pipeline.
	ErrArgsMidPipeline Error = "positional arguments are only allowed on the first command"
	// ErrEmptyCommand is returned for a pipeline segment with no command name.
	ErrEmptyCommand Error = "empty command"
	// ErrFlagNeedsValue is returned when a value-taking flag has no value.
	ErrFlagNeedsValue Error = "flag requires a value"
	// ErrInvalidFlagValue is returned when a flag value is malformed.
	ErrInvalidFlagValue Error = "invalid flag value"
	// ErrInvalidNumber is returned when a numeric argument fails to parse.
	ErrInvalidNumber Error = "invalid number"
	// ErrMissingArgument is returned when a required positional is absent.
	ErrMissingArgument Error = "missing required argument"
	// ErrSourceMidPipeline is returned when a source command (echo, seq, ls,
	// find, yes, emit) appears after a pipe, where only filters are valid.
	ErrSourceMidPipeline Error = "source command cannot appear after a pipe"
	// ErrUnknownCommand is returned for a command name not in the registry.
	ErrUnknownCommand Error = "unknown command"
	// ErrUnknownFlag is returned for a flag not declared by the command.
	ErrUnknownFlag Error = "unknown flag"
	// ErrUnterminatedQuote is returned when a line ends inside a quoted span.
	ErrUnterminatedQuote Error = "unterminated quote"
)
