package engine

import (
	"github.com/jvs-project/jvs/pkg/model"
)

// CloneResult contains the result of a clone operation, including any
// degradation information if the clone could not use the optimal method.
type CloneResult struct {
	Degraded     bool     // true if any degradation occurred
	Degradations []string // list of degradation types
}

// Engine defines the snapshot engine interface for copying worktree data.
type Engine interface {
	// Name returns the engine type identifier.
	Name() model.EngineType

	// Clone performs a copy of src to dst.
	// Returns CloneResult with degradation info if applicable.
	Clone(src, dst string) (*CloneResult, error)
}
