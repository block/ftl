package executor

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
)

type testState struct {
	stage Stage
	kind  ResourceKind
}

func (s *testState) Stage() Stage        { return s.stage }
func (s *testState) Kind() ResourceKind  { return s.kind }
func (s *testState) DebugString() string { return string(s.kind) }

type testExecutor struct {
	stage     Stage
	resources []ResourceKind
	prepared  bool
	executed  bool
}

func (e *testExecutor) Stage() Stage              { return e.stage }
func (e *testExecutor) Resources() []ResourceKind { return e.resources }

func (e *testExecutor) Prepare(ctx context.Context, input State) error {
	e.prepared = true
	return nil
}

func (e *testExecutor) Execute(ctx context.Context) ([]State, error) {
	e.executed = true
	return []State{&testState{StageDone, ResourceKindPostgres}}, nil
}

func TestExecutorPrecedence(t *testing.T) {
	// Create a specific executor for postgres and a default executor
	specificExecutor := &testExecutor{
		stage:     StageProvisioning,
		resources: []ResourceKind{ResourceKindPostgres},
	}
	defaultExecutor := &testExecutor{
		stage:     StageProvisioning,
		resources: AllResources,
	}

	runner := &ProvisionRunner{
		stage:        StageProvisioning,
		CurrentState: []State{&testState{StageProvisioning, ResourceKindPostgres}},
		Executors:    []Executor{defaultExecutor, specificExecutor},
	}

	states, err := runner.Run(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(states))
	assert.Equal(t, StageDone, states[0].Stage())

	assert.True(t, specificExecutor.prepared)
	assert.True(t, specificExecutor.executed)
	assert.False(t, defaultExecutor.prepared)
	assert.False(t, defaultExecutor.executed)
}
