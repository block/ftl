package state

import (
	"context"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/eventstream"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaEvent interface {
	Handle(view SchemaState) (SchemaState, error)
}

type SchemaState struct {
	deployments       map[key.Deployment]*schema.Module
	activeDeployments map[key.Deployment]bool
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, SchemaEvent] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle(&controllerStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state: SchemaState{
			deployments:       map[key.Deployment]*schema.Module{},
			activeDeployments: map[key.Deployment]bool{},
		},
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

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
