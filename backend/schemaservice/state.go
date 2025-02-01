package schemaservice

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/alecthomas/types/optional"
	expmaps "golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	deployments map[key.Deployment]*schema.Module
	// currently active deployments for a given module name. This represents the canonical state of the schema.
	activeDeployments map[string]key.Deployment
	changesets        map[key.Changeset]*ChangesetDetails
}

type ChangesetDetails struct {
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
		changesets:        map[key.Changeset]*ChangesetDetails{},
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

func (r *SchemaState) Marshal() ([]byte, error) {
	activeDeployments := []string{}
	for _, v := range r.activeDeployments {
		activeDeployments = append(activeDeployments, v.String())
	}
	cs := []*schema.SerializedChangeset{}
	for _, v := range r.changesets {
		deps := []string{}
		for _, v := range v.Deployments {
			deps = append(deps, v.String())
		}
		cs = append(cs, &schema.SerializedChangeset{Key: v.Key.String(), CreatedAt: v.CreatedAt, Deployments: deps, State: v.State, Error: v.Error})
	}
	state := &schema.SchemaState{
		Modules:             slices.Collect(maps.Values(r.deployments)),
		ActiveDeployments:   activeDeployments,
		SerializedChangeset: cs,
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
		r.deployments[dkey] = module
	}
	for _, a := range state.ActiveDeployments {
		deploymentKey, err := key.ParseDeploymentKey(a)
		if err != nil {
			return fmt.Errorf("failed to parse deployment key: %w", err)
		}
		r.activeDeployments[deploymentKey.Payload.Module] = deploymentKey
	}
	return nil
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

// GetCanonicalDeployments returns all active deployments (excluding those in changesets).
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
	return expmaps.Values(r.GetCanonicalDeployments())
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
	err := c.state.ApplyEvent(c.runningCtx, msg)
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

func hydrateChangeset(current *SchemaState, changeset *ChangesetDetails) *schema.Changeset {
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
