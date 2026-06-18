package constants

import "fmt"

// Error is the single sentinel error type for the yupsh REPL. Every error the
// program can emit is a constant of this type, so callers test with errors.Is
// rather than comparing strings.
type Error string

// Error satisfies the error interface.
func (e Error) Error() string { return string(e) }

// With wraps an optional cause and optional context arguments around the
// sentinel while keeping errors.Is(result, e) true. It mirrors the framework's
// gloo.Error.With (and gomatic/template.cli's constants.Error.With) so error
// construction stays uniform across the ecosystem.
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

var _ error = Error("")
