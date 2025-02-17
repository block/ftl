package executor

import (
	"context"
	"fmt"
)

type Stage int

const (
	// StageProvisioning is the stage where the infrastructure is being provisioned.
	StageProvisioning Stage = iota
	// StageSetup is the stage where the infrastructure is being setup after provisioning.
	StageSetup Stage = iota
	// StageDone is the stage where the resource is ready to be used.
	StageDone Stage = iota
)

func (s Stage) String() string {
	switch s {
	case StageProvisioning:
		return "provisioning"
	case StageSetup:
		return "setup"
	case StageDone:
		return "done"
	default:
		return fmt.Sprintf("unknown_stage_%d", s)
	}
}

type ResourceKind string

const (
	ResourceKindPostgres ResourceKind = "postgres"
	ResourceKindMySQL    ResourceKind = "mysql"
)

var AllResources []ResourceKind

// State of a single resource in execution.
type State interface {
	Stage() Stage
	Kind() ResourceKind
	DebugString() string
}

type Executor interface {
	// Stage returns the stage that the executor belongs to.
	Stage() Stage
	// Resources returns the resources that the executor supports. Empty means all resources.
	Resources() []ResourceKind
	// Prepare prepares the executor for execution.
	Prepare(ctx context.Context, input State) error
	// Execute executes the executor, returning the next states.
	Execute(ctx context.Context) ([]State, error)
}
