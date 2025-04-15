package schemaservice

import (
	"context"
	"fmt"
	"io"
	"maps"
	"slices"
	"sync"

	"github.com/alecthomas/types/optional"
	expmaps "golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/iterops"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	deployments        map[string]*schema.Module
	changesets         map[key.Changeset]*schema.Changeset
	changesetEvents    map[key.Changeset][]*schema.DeploymentRuntimeEvent
	deploymentEvents   map[string][]*schema.DeploymentRuntimeEvent
	archivedChangesets []*schema.Changeset
	realms             map[string]*schema.RealmState
}

func NewSchemaState() SchemaState {
	return SchemaState{
		deployments:        map[string]*schema.Module{},
		changesets:         map[key.Changeset]*schema.Changeset{},
		deploymentEvents:   map[string][]*schema.DeploymentRuntimeEvent{},
		changesetEvents:    map[key.Changeset][]*schema.DeploymentRuntimeEvent{},
		archivedChangesets: []*schema.Changeset{},
		realms:             map[string]*schema.RealmState{},
	}
}

func NewInMemorySchemaState(ctx context.Context) *statemachine.SingleQueryHandle[struct{}, SchemaState, EventWrapper] {
	handle := statemachine.NewLocalHandle(newStateMachine(ctx))
	return statemachine.NewSingleQueryHandle(handle, struct{}{})
}

func newStateMachine(ctx context.Context) *schemaStateMachine {
	notifier := channels.NewNotifier(ctx)
	return &schemaStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state:      NewSchemaState(),
	}
}

func (r *SchemaState) Marshal() ([]byte, error) {
	if err := r.validate(); err != nil {
		return nil, fmt.Errorf("failed to validate schema state: %w", err)
	}

	changesets := slices.Collect(maps.Values(r.changesets))
	changesets = append(changesets, r.archivedChangesets...)
	dplEvents := []*schema.DeploymentRuntimeEvent{}
	csEvents := []*schema.DeploymentRuntimeEvent{}
	for _, e := range r.changesetEvents {
		csEvents = append(csEvents, e...)
	}
	for _, e := range r.deploymentEvents {
		dplEvents = append(dplEvents, e...)
	}
	state := &schema.SchemaState{
		Modules:          slices.Collect(maps.Values(r.deployments)),
		Changesets:       changesets,
		DeploymentEvents: dplEvents,
		ChangesetEvents:  csEvents,
		Realms:           slices.Collect(maps.Values(r.realms)),
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
	activeDeployments := map[key.Deployment]bool{}
	for _, module := range state.Modules {
		r.deployments[module.Name] = module
		activeDeployments[module.Runtime.Deployment.DeploymentKey] = true
	}
	for _, a := range state.Changesets {
		if a.State == schema.ChangesetStateFinalized || a.State == schema.ChangesetStateFailed {
			r.archivedChangesets = append(r.archivedChangesets, a)
		} else {
			r.changesets[a.Key] = a
		}
	}
	for _, a := range state.DeploymentEvents {
		if activeDeployments[a.DeploymentKey()] {
			r.deploymentEvents[a.DeploymentKey().Payload.Module] = append(r.deploymentEvents[a.DeploymentKey().Payload.Module], a)
		}
	}
	for _, a := range state.ChangesetEvents {
		if cs, ok := a.ChangesetKey().Get(); ok {
			r.changesetEvents[cs] = append(r.changesetEvents[cs], a)
		}
	}
	for _, a := range state.Realms {
		r.realms[a.Name] = a
	}
	return nil
}

func (r *SchemaState) validate() error {
	internals := iterops.Count(maps.Values(r.realms), func(r *schema.RealmState) bool { return !r.External })
	if internals > 1 {
		return fmt.Errorf("only one internal realm is allowed, got %d", internals)
	}
	return nil
}

// GetDeployment returns a deployment based on the deployment key and changeset.
func (r *SchemaState) GetDeployment(deployment key.Deployment, changeset optional.Option[key.Changeset]) (*schema.Module, error) {
	if key, ok := changeset.Get(); ok {
		cs, ok := r.changesets[key]
		if ok {
			modules := cs.OwnedModules(cs.InternalRealm())
			for _, m := range modules {
				if m.GetRuntime().Deployment.DeploymentKey == deployment {
					return m, nil
				}
			}
		}
	}
	d, ok := r.deployments[deployment.Payload.Module]
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
	d, ok := r.deployments[deploymentKey.Payload.Module]
	if ok && d.Runtime.Deployment.DeploymentKey == deploymentKey {
		return d, optional.None[key.Changeset](), nil
	}
	for _, cs := range r.changesets {
		modules := cs.OwnedModules(cs.InternalRealm())
		for _, d := range modules {
			if d.GetRuntime().Deployment.DeploymentKey == deploymentKey {
				return d, optional.Some(cs.Key), nil
			}
		}
	}
	return nil, optional.None[key.Changeset](), fmt.Errorf("deployment %s not found", deploymentKey)
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	//TODO: do we need this method
	ret := map[key.Deployment]*schema.Module{}
	for _, d := range r.deployments {
		ret[d.GetRuntime().Deployment.DeploymentKey] = d
	}
	for _, cs := range r.changesets {
		if cs.ModulesAreCanonical() {
			for _, d := range cs.InternalRealm().Modules {
				ret[d.GetRuntime().Deployment.DeploymentKey] = d
			}
		}
	}
	return ret
}

// GetCanonicalDeployments returns all active deployments (excluding those in changesets).
func (r *SchemaState) GetCanonicalDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for _, dep := range r.deployments {
		deployments[dep.GetRuntime().Deployment.DeploymentKey] = dep
	}
	return deployments
}

// GetCanonicalRealms returns all active realms (excluding those in changesets).
func (r *SchemaState) GetCanonicalRealms() map[string]*schema.RealmState {
	realms := map[string]*schema.RealmState{}
	for _, realm := range r.realms {
		realms[realm.Name] = realm
	}
	return realms
}

// GetCanonicalSchema returns the canonical schema for the active deployments and realms.
func (r *SchemaState) GetCanonicalSchema() *schema.Schema {
	realms := r.GetCanonicalRealms()
	deployments := r.GetCanonicalDeployments()
	realmMap := map[string]*schema.Realm{}

	for _, realm := range realms {
		realmMap[realm.Name] = &schema.Realm{
			Name:     realm.Name,
			External: realm.External,
			Modules:  []*schema.Module{},
		}
	}
	for key, module := range deployments {
		if _, ok := realmMap[key.Payload.Realm]; !ok {
			panic(fmt.Sprintf("realm %s not found for deployment %s", key.Payload.Realm, key.String()))
		}
		realmMap[key.Payload.Realm].Modules = append(realmMap[key.Payload.Realm].Modules, module)
	}

	return &schema.Schema{Realms: expmaps.Values(realmMap)}
}

// GetAllActiveDeployments returns all active deployments, including those in changesets that are prepared
// This includes canary deployments that are not yet committed
func (r *SchemaState) GetAllActiveDeployments() map[key.Deployment]*schema.Module {
	deployments := r.GetCanonicalDeployments()
	for _, cs := range r.changesets {
		if cs.State == schema.ChangesetStatePrepared {
			for _, dep := range cs.InternalRealm().Modules {
				deployments[dep.GetRuntime().Deployment.DeploymentKey] = dep
			}
		}
	}
	return deployments
}

func (r *SchemaState) GetCanonicalDeploymentSchemas() []*schema.Module {
	return expmaps.Values(r.GetCanonicalDeployments())
}

func (r *SchemaState) GetProvisioning(module string, cs key.Changeset) (*schema.Module, error) {
	c, ok := r.changesets[cs]
	if !ok {
		return nil, fmt.Errorf("changeset %s not found", cs.String())
	}
	for _, m := range c.InternalRealm().Modules {
		if m.Name == module {
			return m, nil
		}
	}
	return nil, fmt.Errorf("provisioning for module %s not found", module)
}

type EventWrapper struct {
	Event schema.Event
}

func (e EventWrapper) String() string {
	return fmt.Sprintf("EventWrapper{Event: %T}", e.Event)
}

var _ statemachine.Marshallable = EventWrapper{}

func (e EventWrapper) MarshalBinary() ([]byte, error) {
	pb := schema.EventToProto(e.Event)
	bytes, err := proto.Marshal(pb)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event: %w", err)
	}
	return bytes, nil
}

func (e *EventWrapper) UnmarshalBinary(bts []byte) error {
	pb := schemapb.Event{}
	if err := proto.Unmarshal(bts, &pb); err != nil {
		return fmt.Errorf("error unmarshalling event proto: %w", err)
	}
	event, err := schema.EventFromProto(&pb)
	if err != nil {
		return fmt.Errorf("error decoding event proto: %w", err)
	}
	e.Event = event
	return nil
}

type schemaStateMachine struct {
	state SchemaState

	notifier   *channels.Notifier
	runningCtx context.Context

	lock sync.Mutex
}

var _ statemachine.Snapshotting[struct{}, SchemaState, EventWrapper] = &schemaStateMachine{}
var _ statemachine.Listenable[struct{}, SchemaState, EventWrapper] = &schemaStateMachine{}

func (c *schemaStateMachine) Lookup(key struct{}) (SchemaState, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return reflect.DeepCopy(c.state), nil
}

func (c *schemaStateMachine) Publish(msg EventWrapper) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	logger := log.FromContext(c.runningCtx)

	err := c.state.ApplyEvent(c.runningCtx, msg.Event)
	if err != nil {
		// TODO: we need to validate the events before they are
		// committed to the log
		logger.Errorf(err, "failed to apply event")
		return raft.ErrInvalidEvent
	}
	// Notify all subscribers using broadcaster
	c.notifier.Notify(c.runningCtx)
	return nil
}

func (c *schemaStateMachine) Subscribe(ctx context.Context) (<-chan struct{}, error) {
	return c.notifier.Subscribe(ctx), nil
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
