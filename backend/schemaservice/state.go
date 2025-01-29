package schemaservice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	// TODO: updating this info is very tricky as both the module and the changeset need to be updated
	deployments map[key.Deployment]*schema.Module
	// currently active deployments for a given module name. This represents the canonical state of the schema.
	activeDeployments map[string]key.Deployment
	changesets        map[key.Changeset]*changesetDetails
	// TODO: consider removing committed changesets. Return success if asked about a missing changeset? Or keep the shell of changesets

	provisioning map[string]key.Deployment
}

type changesetDetails struct {
	Key         key.Changeset
	CreatedAt   time.Time
	Deployments []key.Deployment
	State       schema.ChangesetState
	// Error is present if state is failed.
	Error string
}

func NewSchemaState() SchemaState {
	return SchemaState{
		deployments:       map[key.Deployment]*schema.Module{},
		activeDeployments: map[string]key.Deployment{},
		changesets:        map[key.Changeset]*changesetDetails{},
		provisioning:      map[string]key.Deployment{},
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
	//TODO: remove this
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
	return nil, optional.None[key.Changeset](), fmt.Errorf("deployment %s not found", deploymentKey)
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	return r.deployments
}

// GetActiveDeployments returns all active deployments (excluding those in changesets).
func (r *SchemaState) GetCanonicalDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for _, dep := range r.activeDeployments {
		deployments[dep] = r.deployments[dep]
	}
	return deployments
}

// GetAllActiveDeployments returns all active deployments, including those in changesets.
func (r *SchemaState) GetAllActiveDeployments() map[key.Deployment]*schema.Module {
	deployments := r.GetCanonicalDeployments()
	for _, cs := range r.changesets {
		for _, dep := range cs.Deployments {
			deployments[dep] = r.deployments[dep]
		}
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
	return r.deployments[d], nil
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
	c.state, err = c.state.ApplyEvent(c.runningCtx, msg)
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

func hydrateChangeset(current *SchemaState, changeset *changesetDetails) *schema.Changeset {
	changesetModules := make([]*schema.Module, len(changeset.Deployments))
	for i, deployment := range changeset.Deployments {
		changesetModules[i] = current.deployments[deployment]
	}
	return &schema.Changeset{
		Key:       changeset.Key,
		CreatedAt: changeset.CreatedAt,
		State:     changeset.State,
		Modules:   changesetModules,
		Error:     changeset.Error,
	}
}
