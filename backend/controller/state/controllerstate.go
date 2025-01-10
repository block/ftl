package state

import (
	"context"

	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaEvent interface {
	Handle(view SchemaState) (SchemaState, error)
}

type SchemaState struct {
	deployments       map[string]*Deployment
	activeDeployments map[string]bool
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, SchemaEvent] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle(&controllerStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state: SchemaState{
			deployments:       map[string]*Deployment{},
			activeDeployments: map[string]bool{},
		},
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

type RunnerState struct {
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
}

type RunnerEvent interface {
	Handle(view RunnerState) (RunnerState, error)
}

func NewInMemoryRunnerState(ctx context.Context) eventstream.EventStream[RunnerState, RunnerEvent] {
	return eventstream.NewInMemory[RunnerState, RunnerEvent](RunnerState{
		runners:             map[string]*Runner{},
		runnersByDeployment: map[string][]*Runner{},
	})
}
