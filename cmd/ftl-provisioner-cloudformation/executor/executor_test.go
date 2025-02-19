package executor

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/internal/log"
)

type testState struct{}

func (s *testState) DebugString() string { return "" }

type ignoredState struct{}

func (s *ignoredState) DebugString() string { return "" }

type testExecutor struct {
	prepared int
	executed int
}

func (e *testExecutor) Prepare(ctx context.Context, input State) error {
	e.prepared++
	return nil
}

func (e *testExecutor) Execute(ctx context.Context) ([]State, error) {
	e.executed++
	if e.prepared > 0 {
		return []State{&testState{}}, nil
	}
	return nil, nil
}

func TestExecutorSelection(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	postgresExecutor := &testExecutor{}
	mysqlExecutor := &testExecutor{}

	runner := &ProvisionRunner{
		CurrentState: []State{&testState{}, &testState{}, &ignoredState{}},
		Stages: []RunnerStage{
			{
				Executors: []Handler{
					{postgresExecutor, []State{&testState{}}},
					{mysqlExecutor, []State{}},
				},
			},
		},
	}

	states, err := runner.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(states))

	assert.Equal(t, 2, postgresExecutor.prepared)
	assert.Equal(t, 1, postgresExecutor.executed)
	assert.Equal(t, 0, mysqlExecutor.prepared)
	assert.Equal(t, 1, mysqlExecutor.executed)
}
