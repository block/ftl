package state

import (
	"context"
	"fmt"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/statemachine"
)

type controllerStateMachine struct {
	state SchemaState

	notifier   *channels.Notifier
	runningCtx context.Context
}

var _ statemachine.Listenable[struct{}, SchemaState, SchemaEvent] = &controllerStateMachine{}

func (c *controllerStateMachine) Lookup(key struct{}) (SchemaState, error) {
	return reflect.DeepCopy(c.state), nil
}

func (c *controllerStateMachine) Publish(msg SchemaEvent) error {
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
