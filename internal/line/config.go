package line

import (
	"github.com/spf13/afero"

	"github.com/yupsh/repl/internal/expansion"
)

// Config holds the injected collaborators a line needs to be resolved and
// planned: the filesystem (for glob matching, file sources, and ls/find
// construction) and the home directory (for "~" expansion). Both are injected so
// resolution stays testable against an in-memory afero.Fs. It carries data only,
// no behavior.
//
// There is no domain-local types.go: Fs reuses afero.Fs and Home reuses
// expansion.Home, so there are no bare primitives to name.
type Config struct {
	Fs   afero.Fs
	Home expansion.Home
}
