package provisioner

import (
	"context"

	"github.com/block/ftl/internal/provisioner/state"
)

// Executor executes a single step in provisioner execution.
// It transforms a set of states into their next stages.
type Executor interface {
	// Stage returns the stage that the executor belongs to.
	// Prepare prepares the executor for execution.
	Prepare(ctx context.Context, input state.State) error
	// Execute executes the executor, returning the next states.
	Execute(ctx context.Context) ([]state.State, error)
}
