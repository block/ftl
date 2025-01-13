package state

import (
	"context"

	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaEvent interface {
	Handle(view SchemaState) (SchemaState, error)
}

type SchemaState struct {
	deployments       map[model.DeploymentKey]*Deployment
	activeDeployments map[model.DeploymentKey]bool
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, SchemaEvent] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle(&controllerStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state: SchemaState{
			deployments:       map[model.DeploymentKey]*Deployment{},
			activeDeployments: map[model.DeploymentKey]bool{},
		},
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

type RunnerState struct {
	runners             map[model.RunnerKey]*Runner
	runnersByDeployment map[model.DeploymentKey][]*Runner
}

type RunnerEvent interface {
	Handle(view RunnerState) (RunnerState, error)
}

func NewInMemoryRunnerState(ctx context.Context) eventstream.EventStream[RunnerState, RunnerEvent] {
	return eventstream.NewInMemory[RunnerState, RunnerEvent](RunnerState{
		runners:             map[model.RunnerKey]*Runner{},
		runnersByDeployment: map[model.DeploymentKey][]*Runner{},
	})
}
