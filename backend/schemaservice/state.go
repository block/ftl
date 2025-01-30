package schemaservice

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
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

func (r *SchemaState) Marshal() ([]byte, error) {
	state := &schema.SchemaState{
		Modules: append(slices.Collect(maps.Values(r.deployments)), slices.Collect(maps.Values(r.provisioning))...),
	}
	stateProto := state.ToProto()
	bytes, err := proto.Marshal(stateProto)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema state: %w", err)
	}
	return bytes, nil
}

func (r *SchemaState) Unmarshal(data []byte) error {
	stateProto := &schemapb.SchemaState{}
	if err := proto.Unmarshal(data, stateProto); err != nil {
		return fmt.Errorf("failed to unmarshal schema state: %w", err)
	}

	state, err := schema.SchemaStateFromProto(stateProto)
	if err != nil {
		return fmt.Errorf("failed to unmarshal schema state: %w", err)
	}

	for _, module := range state.Modules {
		dkey := module.GetRuntime().GetDeployment().GetDeploymentKey()
		if dkey.IsZero() {
			r.provisioning[module.Name] = module
		} else {
			r.deployments[dkey] = module
			if module.GetRuntime().GetDeployment().ActivatedAt.Ok() {
				r.activeDeployments[dkey] = true
			}
		}
	}

	return nil
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

var _ statemachine.Snapshotting[struct{}, SchemaState, schema.Event] = &schemaStateMachine{}
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

func (c *schemaStateMachine) Close() error {
	return nil
}

func (c *schemaStateMachine) Recover(snapshot io.Reader) error {
	snapshotBytes, err := io.ReadAll(snapshot)
	if err != nil {
		return fmt.Errorf("failed to read snapshot: %w", err)
	}
	if err := c.state.Unmarshal(snapshotBytes); err != nil {
		return fmt.Errorf("failed to unmarshal snapshot: %w", err)
	}
	return nil
}

func (c *schemaStateMachine) Save(w io.Writer) error {
	snapshotBytes, err := c.state.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}
	_, err = w.Write(snapshotBytes)
	if err != nil {
		return fmt.Errorf("failed to write snapshot: %w", err)
	}
	return nil
}
