package repl

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorWith(t *testing.T) {
	const sentinel Error = "boom"
	cause := errors.New("cause")

	if err := sentinel.With(nil); !errors.Is(err, sentinel) {
		t.Errorf("With(nil) lost sentinel: %v", err)
	}

	wrapped := sentinel.With(cause)
	if !errors.Is(wrapped, sentinel) || !errors.Is(wrapped, cause) {
		t.Errorf("With(cause) lost a layer: %v", wrapped)
	}

	ctx := sentinel.With(nil, "detail")
	if !errors.Is(ctx, sentinel) || !strings.Contains(ctx.Error(), "detail") {
		t.Errorf("With(nil, detail) = %v", ctx)
	}

	full := sentinel.With(cause, "detail")
	if !errors.Is(full, sentinel) || !errors.Is(full, cause) || !strings.Contains(full.Error(), "detail") {
		t.Errorf("With(cause, detail) = %v", full)
	}
}

func TestErrorString(t *testing.T) {
	if ErrUnknownCommand.Error() != "unknown command" {
		t.Errorf("Error() = %q", ErrUnknownCommand.Error())
	}
}
