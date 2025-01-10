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
	activeDeployments   map[string]bool
	runners             map[string]*Runner
	runnersByDeployment map[string][]*Runner
}

type ControllerEvent interface {
	Handle(view State) (State, error)
}

type controllerStateMachine struct {
	state State

	notifier   *channels.Notifier
	runningCtx context.Context
}

var _ statemachine.Listenable[struct{}, State, ControllerEvent] = &controllerStateMachine{}

func (c *controllerStateMachine) Lookup(key struct{}) (State, error) {
	return reflect.DeepCopy(c.state), nil
}

func (c *controllerStateMachine) Publish(msg ControllerEvent) error {
	var err error
	c.state, err = msg.Handle(c.state)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	// Notify all subscribers using broadcaster
	c.notifier.Notify(c.runningCtx)
	return nil
}

func (c *controllerStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return c.notifier.Subscribe(), nil
}

func NewInMemoryState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, State, ControllerEvent] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle(&controllerStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state: State{
			deployments:         map[string]*Deployment{},
			activeDeployments:   map[string]bool{},
			runners:             map[string]*Runner{},
			runnersByDeployment: map[string][]*Runner{},
		},
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}
