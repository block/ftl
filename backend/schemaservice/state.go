package schemaservice

import (
	"context"
	"fmt"
	"sync"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	// TODO: updating this info is very tricky as both the module and the changeset need to be updated
	deployments       map[key.Deployment]*schema.Module
	activeDeployments map[key.Deployment]bool
	changesets        map[key.Changeset]*schema.Changeset
	// TODO: consider removing committed changesets. Return success if asked about a missing changeset? Or keep the shell of changesets

	provisioning map[string]*schema.Module
}

func NewSchemaState() SchemaState {
	return SchemaState{
		deployments:       map[key.Deployment]*schema.Module{},
		activeDeployments: map[key.Deployment]bool{},
		changesets:        map[key.Changeset]*schema.Changeset{},
		provisioning:      map[string]*schema.Module{},
	}
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, schema.Event] {
	notifier := channels.NewNotifier(ctx)
	handle := statemachine.NewLocalHandle[struct{}, SchemaState, schema.Event](&schemaStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state:      NewSchemaState(),
	})

	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

// GetDeployment returns a deployment based on the deployment key and changeset.
func (r *SchemaState) GetDeployment(deployment key.Deployment, changeset optional.Option[key.Changeset]) (*schema.Module, error) {
	if changesetKey, ok := changeset.Get(); ok {
		c, ok := r.changesets[changesetKey]
		if !ok {
			return nil, fmt.Errorf("changeset %s not found", changesetKey)
		}
		dep, ok := slices.Find(c.Modules, func(m *schema.Module) bool {
			return m.Runtime.Deployment.DeploymentKey == deployment
		})
		if !ok {
			return nil, fmt.Errorf("deployment %s not found in changeset %s", deployment, changesetKey)
		}
		return dep, nil
	}
	d, ok := r.deployments[deployment]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", deployment)
	}
	return d, nil
}

// FindDeployment returns a deployment and which changeset it is in based on the deployment key.
func (r *SchemaState) FindDeployment(deploymentKey key.Deployment) (deployment *schema.Module, changeset optional.Option[key.Changeset], err error) {
	// TODO: add unit tests:
	// - deployment in ended changeset + canonical
	// - deployment in ended changeset + main list but no longer canonical
	// - deployment in failed changeset
	d, ok := r.deployments[deploymentKey]
	if ok {
		return d, optional.None[key.Changeset](), nil
	}
	for _, c := range r.changesets {
		for _, m := range c.Modules {
			if m.Runtime.Deployment.DeploymentKey == deploymentKey {
				return m, optional.Some(c.Key), nil
			}
		}
	}
	return nil, optional.None[key.Changeset](), fmt.Errorf("deployment %s not found", deploymentKey)
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	return r.deployments
}

// GetActiveDeployments returns all active deployments (excluding those in changesets).
func (r *SchemaState) GetCanonicalDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for key, _ := range r.activeDeployments {
		deployments[key] = r.deployments[key]
	}
	return deployments
}

// GetAllActiveDeployments returns all active deployments, including those in changesets.
func (r *SchemaState) GetAllActiveDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for key := range r.activeDeployments {
		deployments[key] = r.deployments[key]
	}
	return deployments
}

func (r *SchemaState) GetCanonicalDeploymentSchemas() []*schema.Module {
	return maps.Values(r.GetCanonicalDeployments())
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

type moduleRef struct {
	module *schema.Module
}
