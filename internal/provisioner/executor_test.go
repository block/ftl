package provisioner

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner/state"
)

type testState struct{}

func (s *testState) DebugString() string { return "" }

type ignoredState struct{}

func (s *ignoredState) DebugString() string { return "" }

type testExecutor struct {
	prepared int
	executed int
}

func (e *testExecutor) Prepare(ctx context.Context, input state.State) error {
	e.prepared++
	return nil
}

func (e *testExecutor) Execute(ctx context.Context) ([]state.State, error) {
	e.executed++
	if e.prepared > 0 {
		return []state.State{&testState{}}, nil
	}
	return nil, nil
}

func TestExecutorSelection(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	postgresExecutor := &testExecutor{}
	mysqlExecutor := &testExecutor{}

	runner := &Runner{
		State: []state.State{&testState{}, &testState{}, &ignoredState{}},
		Stages: []RunnerStage{
			{
				Handlers: []Handler{
					{postgresExecutor, []state.State{&testState{}}},
					{mysqlExecutor, []state.State{}},
				},
			},
		},
	}

	states, err := runner.Run(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(states))

	assert.Equal(t, 2, postgresExecutor.prepared)
	assert.Equal(t, 1, postgresExecutor.executed)
	assert.Equal(t, 0, mysqlExecutor.prepared)
	assert.Equal(t, 1, mysqlExecutor.executed)
}
