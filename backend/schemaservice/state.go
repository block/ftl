package schemaservice

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	expmaps "golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	ftlslices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/maps"
	"github.com/block/ftl/internal/raft"
	"github.com/block/ftl/internal/statemachine"
)

type SchemaState struct {
	state              *schema.SchemaState
	archivedChangesets []*schema.Changeset
}

func NewSchemaState(canonicalRealm string) SchemaState {
	sch := &schema.Schema{
		Realms: []*schema.Realm{},
	}
	if canonicalRealm != "" {
		sch.Realms = append(sch.Realms, &schema.Realm{
			Name:     canonicalRealm,
			External: false,
		})
	}
	return SchemaState{
		state: &schema.SchemaState{
			Schema:           sch,
			Changesets:       []*schema.Changeset{},
			DeploymentEvents: []*schema.DeploymentRuntimeEvent{},
			ChangesetEvents:  []*schema.DeploymentRuntimeEvent{},
		},
	}
}

func newStateMachine(ctx context.Context, realm string) *schemaStateMachine {
	notifier := channels.NewNotifier(ctx)
	return &schemaStateMachine{
		notifier:   notifier,
		runningCtx: ctx,
		state:      NewSchemaState(realm),
	}
}

func (r *SchemaState) Marshal() ([]byte, error) {
	if err := r.state.Validate(); err != nil {
		return nil, errors.Wrap(err, "failed to validate schema state")
	}
	bytes, err := proto.Marshal(r.state.ToProto())
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal schema state")
	}
	return bytes, nil
}

func (r *SchemaState) Unmarshal(data []byte) error {
	stateProto := &schemapb.SchemaState{}
	if err := proto.Unmarshal(data, stateProto); err != nil {
		return errors.Wrap(err, "failed to unmarshal schema state")
	}
	state, err := schema.SchemaStateFromProto(stateProto)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal schema state")
	}
	r.state = state
	return nil
}

// FindDeployment returns a deployment and which changeset it is in based on the deployment key.
func (r *SchemaState) FindDeployment(deploymentKey key.Deployment) (deployment *schema.Module, changeset optional.Option[key.Changeset], err error) {
	// TODO: add unit tests:
	// - deployment in ended changeset + canonical
	// - deployment in ended changeset + main list but no longer canonical
	// - deployment in failed changeset

	for _, realm := range r.state.Schema.Realms {
		for _, module := range realm.Modules {
			if module.GetRuntime() == nil || module.GetRuntime().Deployment == nil {
				continue
			}
			if module.GetRuntime().Deployment.DeploymentKey == deploymentKey {
				return module, optional.None[key.Changeset](), nil
			}
		}
	}

	for _, cs := range r.state.Changesets {
		modules := cs.OwnedModules(cs.InternalRealm())
		for _, d := range modules {
			if d.GetRuntime().Deployment.DeploymentKey == deploymentKey {
				return d, optional.Some(cs.Key), nil
			}
		}
	}
	return nil, optional.None[key.Changeset](), errors.Errorf("deployment %s not found", deploymentKey)
}

func (r *SchemaState) GetDeployments() map[key.Deployment]*schema.Module {
	ret := map[key.Deployment]*schema.Module{}
	if r.state.Schema != nil {
		for _, realm := range r.state.Schema.Realms {
			for _, d := range realm.Modules {
				if d.GetRuntime() == nil || d.GetRuntime().Deployment == nil {
					continue
				}
				ret[d.GetRuntime().Deployment.DeploymentKey] = d
			}
		}
	}
	if r.state.Changesets != nil {
		for _, cs := range r.state.Changesets {
			for _, m := range cs.InternalRealm().Modules {
				d := m.GetRuntime().GetDeployment()
				if d == nil {
					continue
				}
				ret[d.DeploymentKey] = m
			}
		}
	}
	return ret
}

func (r *SchemaState) RealmsByName() map[string]*schema.Realm {
	return maps.FromSlice(r.state.Schema.Realms, realmByName)
}

// GetCanonicalDeployments returns all active deployments (excluding those in changesets).
func (r *SchemaState) GetCanonicalDeployments() map[key.Deployment]*schema.Module {
	deployments := map[key.Deployment]*schema.Module{}
	for _, realm := range r.state.Schema.Realms {
		for _, dep := range realm.Modules {
			if dep.GetRuntime() == nil || dep.GetRuntime().Deployment == nil {
				continue
			}
			deployments[dep.GetRuntime().Deployment.DeploymentKey] = dep
		}
	}
	return deployments
}

// GetCanonicalRealms returns all active realms (excluding those in changesets).
func (r *SchemaState) GetCanonicalRealms() map[string]*schema.Realm {
	realms := map[string]*schema.Realm{}
	for _, realm := range r.state.Schema.Realms {
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

func (r *SchemaState) ChangesetEvents(key key.Changeset) []*schema.DeploymentRuntimeEvent {
	events := []*schema.DeploymentRuntimeEvent{}
	for _, event := range r.state.ChangesetEvents {
		if event.Changeset == nil {
			continue
		}
		if *event.Changeset == key {
			events = append(events, event)
		}
	}
	return events
}

func (r *SchemaState) DeploymentEvents(module string) []*schema.DeploymentRuntimeEvent {
	events := []*schema.DeploymentRuntimeEvent{}
	for _, event := range r.state.DeploymentEvents {
		if event.Payload.Deployment.Payload.Module == module {
			events = append(events, event)
		}
	}
	return events
}

func (r *SchemaState) deleteChangeset(key key.Changeset) {
	r.state.Changesets = ftlslices.Filter(r.state.Changesets, func(cs *schema.Changeset) bool {
		return cs.Key != key
	})
	r.state.ChangesetEvents = ftlslices.Filter(r.state.ChangesetEvents, func(event *schema.DeploymentRuntimeEvent) bool {
		return event.Changeset == nil || *event.Changeset != key
	})
}

func (r *SchemaState) upsertChangeset(cs *schema.Changeset) {
	for i, c := range r.state.Changesets {
		if c.Key == cs.Key {
			r.state.Changesets[i] = cs
			return
		}
	}
	r.state.Changesets = append(r.state.Changesets, cs)
}

func (r *SchemaState) clearDeploymentEvents(module string) {
	r.state.DeploymentEvents = ftlslices.Filter(r.state.DeploymentEvents, func(event *schema.DeploymentRuntimeEvent) bool {
		return event.Payload.Deployment.Payload.Module != module
	})
}

type NiceSlice[T any] []T

func (s NiceSlice[T]) Filter(f func(T) bool) NiceSlice[T] {
	return ftlslices.Filter(s, f)
}

func (s NiceSlice[T]) Map(f func(T) T) NiceSlice[T] {
	return ftlslices.Map(s, f)
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
		return nil, errors.Wrap(err, "failed to marshal event")
	}
	return bytes, nil
}

func (e *EventWrapper) UnmarshalBinary(bts []byte) error {
	pb := schemapb.Event{}
	if err := proto.Unmarshal(bts, &pb); err != nil {
		return errors.Wrap(err, "error unmarshalling event proto")
	}
	event, err := schema.EventFromProto(&pb)
	if err != nil {
		return errors.Wrap(err, "error decoding event proto")
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
		return errors.WithStack(raft.ErrInvalidEvent)
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
		return errors.Wrap(err, "failed to read snapshot")
	}
	if err := c.state.Unmarshal(snapshotBytes); err != nil {
		return errors.Wrap(err, "failed to unmarshal snapshot")
	}
	return nil
}

func (c *schemaStateMachine) Save(w io.Writer) error {
	snapshotBytes, err := c.state.Marshal()
	if err != nil {
		return errors.Wrap(err, "failed to marshal snapshot")
	}
	_, err = w.Write(snapshotBytes)
	if err != nil {
		return errors.Wrap(err, "failed to write snapshot")
	}
	return nil
}

func realmByName(m *schema.Realm) (string, *schema.Realm) { return m.Name, m }
