package state

import (
	"context"

	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/key"
)

type RunnerState struct {
	runners             map[key.Runner]*Runner
	runnersByDeployment map[key.Deployment][]*Runner
}

type RunnerEvent interface {
	Handle(view RunnerState) (RunnerState, error)
}

func NewInMemoryRunnerState(ctx context.Context) eventstream.EventStream[RunnerState, RunnerEvent] {
	return eventstream.NewInMemory[RunnerState, RunnerEvent](RunnerState{
		runners:             map[key.Runner]*Runner{},
		runnersByDeployment: map[key.Deployment][]*Runner{},
	})
}
