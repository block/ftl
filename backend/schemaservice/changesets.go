package schemaservice

import (
	"fmt"

	"github.com/alecthomas/types/optional"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func (r *SchemaState) ActiveChangeset() optional.Option[*schema.Changeset] {
	for _, changeset := range r.changesets {
		if changeset.State == schema.ChangesetStateProvisioning {
			return optional.Some(changeset)
		}
	}
	return optional.None[*schema.Changeset]()
}

func (r *SchemaState) GetChangeset(changeset key.Changeset) (*schema.Changeset, error) {
	c, ok := r.changesets[changeset]
	if !ok {
		return nil, fmt.Errorf("changeset %s not found", changeset)
	}
	return c, nil
}

func (r *SchemaState) GetChangesets() map[key.Changeset]*schema.Changeset {
	return r.changesets
}

var _ SchemaEvent = (*ChangesetCreatedEvent)(nil)
var _ SchemaEvent = (*ChangesetCommittedEvent)(nil)
var _ SchemaEvent = (*ChangesetFailedEvent)(nil)

type ChangesetCreatedEvent struct {
	Changeset *schema.Changeset
}

func (e *ChangesetCreatedEvent) Handle(t SchemaState) (SchemaState, error) {
	if existing := t.changesets[e.Changeset.Key]; existing != nil {
		return t, nil
	}
	if e.Changeset.State == schema.ChangesetStateProvisioning {
		if active, ok := t.ActiveChangeset().Get(); ok {
			// TODO: make unit test for this
			// TODO: how does error handling work here? Does the changeset need to be added but immediately failed? Or is this error propagated to the caller?
			return t, fmt.Errorf("can not create active changeset: %s already active", active.Key)
		}
	}
	t.changesets[e.Changeset.Key] = e.Changeset
	return t, nil
}

type ChangesetCommittedEvent struct {
	Key key.Changeset
}

func (e *ChangesetCommittedEvent) Handle(t SchemaState) (SchemaState, error) {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return SchemaState{}, fmt.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateCommitted
	for _, module := range changeset.Modules {
		t.deployments[module.Runtime.Deployment.DeploymentKey] = module
	}
	return t, nil
}

type ChangesetFailedEvent struct {
	Key   key.Changeset
	Error string
}

func (e *ChangesetFailedEvent) Handle(t SchemaState) (SchemaState, error) {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return SchemaState{}, fmt.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateFailed
	changeset.Error = e.Error
	return t, nil
}
