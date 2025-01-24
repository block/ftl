package schemaservice

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	deployments       map[key.Deployment]*schema.Module
	activeDeployments map[key.Deployment]bool
	// modules being provisioned
	provisioning map[string]*schema.Module
}

func NewSchemaState() SchemaState {
	return SchemaState{
		deployments:       map[key.Deployment]*schema.Module{},
		activeDeployments: map[key.Deployment]bool{},
		provisioning:      map[string]*schema.Module{},
	}
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, schema.Event] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle(&schemaStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state:      NewSchemaState(),
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

func (r *SchemaState) GetDeployment(deployment key.Deployment) (*schema.Module, error) {
	d, ok := r.deployments[deployment]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	return r.deployments
}

func (r *SchemaState) GetActiveDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for key, active := range r.activeDeployments {
		if active {
			deployments[key] = r.deployments[key]
		}
	}
	return deployments
}

func (r *SchemaState) GetActiveDeploymentSchemas() []*schema.Module {
	return slices.Collect(maps.Values(r.GetActiveDeployments()))
}

func (r *SchemaState) GetProvisioning(moduleName string) (*schema.Module, error) {
	d, ok := r.provisioning[moduleName]
	if !ok {
		return nil, fmt.Errorf("provisioning for module %s not found", moduleName)
	}
	return d, nil
}

type schemaStateMachine struct {
	state SchemaState

	notifier   *channels.Notifier
	runningCtx context.Context

	lock sync.Mutex
}

var _ statemachine.Listenable[struct{}, SchemaState, schema.Event] = &schemaStateMachine{}

func (c *schemaStateMachine) Lookup(key struct{}) (SchemaState, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return reflect.DeepCopy(c.state), nil
}

func (c *schemaStateMachine) Publish(msg schema.Event) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	var err error
	c.state, err = c.state.ApplyEvent(msg)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	// Notify all subscribers using broadcaster
	c.notifier.Notify(c.runningCtx)
	return nil
}

func (c *schemaStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return c.notifier.Subscribe(), nil
}
