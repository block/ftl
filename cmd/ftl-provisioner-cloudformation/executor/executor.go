package executor

import (
	"context"
)

type ResourceKind string

const (
	ResourceKindPostgres ResourceKind = "postgres"
	ResourceKindMySQL    ResourceKind = "mysql"
)

var AllResources []ResourceKind

// State of a single resource in execution.
type State interface {
	DebugString() string
}

type Executor interface {
	// Stage returns the stage that the executor belongs to.
	// Prepare prepares the executor for execution.
	Prepare(ctx context.Context, input State) error
	// Execute executes the executor, returning the next states.
	Execute(ctx context.Context) ([]State, error)
}
