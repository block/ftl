package state

import (
	"context"
	"fmt"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/statemachine"
)

type State struct {
	deployments         map[string]*Deployment
	activeDeployments   map[string]*Deployment
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
}

type ControllerEvent interface {
	Handle(view State) (State, error)
}

type ControllerState statemachine.StateMachineHandle[struct{}, State, ControllerEvent]

type controllerStateMachine struct {
	state State

	broadcaster *channels.Broadcaster[struct{}]
	runningCtx  context.Context
}

var _ statemachine.ListenableStateMachine[struct{}, State, ControllerEvent] = &controllerStateMachine{}

func (c *controllerStateMachine) Lookup(key struct{}) (State, error) {
	return reflect.DeepCopy(c.state), nil
}

func (c *controllerStateMachine) Update(msg ControllerEvent) error {
	var err error
	c.state, err = msg.Handle(c.state)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	// Notify all subscribers using broadcaster
	c.broadcaster.Broadcast(c.runningCtx, struct{}{})
	return nil
}

func (c *controllerStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return c.broadcaster.Subscribe(), nil
}

func NewInMemoryState(ctx context.Context) ControllerState {
	broadcaster := &channels.Broadcaster[struct{}]{}
	go broadcaster.Run(ctx)

	return statemachine.LocalHandle(&controllerStateMachine{
		broadcaster: broadcaster,
		runningCtx:  ctx,
		state: State{
			deployments:         map[string]*Deployment{},
			activeDeployments:   map[string]*Deployment{},
			runners:             map[string]*Runner{},
			runnersByDeployment: map[string][]*Runner{},
		},
	})
}
