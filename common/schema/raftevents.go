package schema

import (
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

// TODO: these should be moved to the schema service package once go2proto supports referring to proto messages from other packages

//sumtype:decl
//protobuf:export
type Event interface {
	event()
	// Validate the event is internally consistent
	Validate() error
	// DebugString returns a string representation of the event for debugging purposes
	DebugString() string
}

// deployment events
var _ Event = (*DeploymentCreatedEvent)(nil)
var _ Event = (*DeploymentRuntimeEvent)(nil)

var _ Event = (*ChangesetCreatedEvent)(nil)
var _ Event = (*ChangesetPreparedEvent)(nil)
var _ Event = (*ChangesetCommittedEvent)(nil)
var _ Event = (*ChangesetFailedEvent)(nil)

//protobuf:1
type DeploymentCreatedEvent struct {
	Key       key.Deployment `protobuf:"1"`
	Schema    *Module        `protobuf:"2"`
	Changeset *key.Changeset `protobuf:"3"`
}

func (r *DeploymentCreatedEvent) DebugString() string {
	return fmt.Sprintf("DeploymentCreatedEvent{key: %s, changeset: %s}", r.Key.String(), r.Changeset.String())
}

func (r *DeploymentCreatedEvent) event() {}

func (r *DeploymentCreatedEvent) Validate() error {
	if r.Schema.GetRuntime().GetDeployment().GetDeploymentKey().IsZero() {
		return fmt.Errorf("deployment key is required")
	}
	return nil
}

//protobuf:2
type DeploymentRuntimeEvent struct {
	Payload   *RuntimeElement `protobuf:"1"`
	Changeset *key.Changeset  `protobuf:"2"`
}

func (e *DeploymentRuntimeEvent) DeploymentKey() key.Deployment {
	return e.Payload.Deployment
}

func (e *DeploymentRuntimeEvent) ChangesetKey() optional.Option[key.Changeset] {
	return optional.Ptr(e.Changeset)
}

func (e *DeploymentRuntimeEvent) DebugString() string {
	return fmt.Sprintf("DeploymentRuntimeEvent{module: %s, changeset: %s, id: %v, data: %v}", e.DeploymentKey().String(), e.Changeset.String(), e.Payload.Name, e.Payload.Element)
}

func (e *DeploymentRuntimeEvent) event() {}

func (e *DeploymentRuntimeEvent) Validate() error {
	return nil
}

//protobuf:3
type ChangesetCreatedEvent struct {
	Changeset *Changeset `protobuf:"1"`
}

func (e *ChangesetCreatedEvent) DebugString() string {
	ret := fmt.Sprintf("ChangesetCreatedEvent{key: %s", e.Changeset.Key.String())
	for _, m := range e.Changeset.Modules {
		if m.Runtime == nil || m.Runtime.Deployment == nil {
			ret += fmt.Sprintf(", invalid module: %s", m.Name)
		} else {
			ret += fmt.Sprintf(", deployment: %s", m.Runtime.Deployment.DeploymentKey.String())
		}
	}
	return ret + "}"
}

func (e *ChangesetCreatedEvent) event() {}

func (e *ChangesetCreatedEvent) Validate() error {
	if e.Changeset.Key.IsZero() {
		return fmt.Errorf("changeset key is required")
	}
	return nil
}

//protobuf:4
type ChangesetPreparedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetPreparedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetPreparedEvent{changeset: %s}", e.Key.String())
}

func (e *ChangesetPreparedEvent) event() {}

func (e *ChangesetPreparedEvent) Validate() error {
	return nil
}

//protobuf:5
type ChangesetCommittedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetCommittedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetCommittedEvent{changeset: %s}", e.Key.String())
}

func (e *ChangesetCommittedEvent) event() {}

func (e *ChangesetCommittedEvent) Validate() error {
	return nil
}

//protobuf:6
type ChangesetDrainedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetDrainedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetDrainedEvent{changeset: %s}", e.Key.String())
}

func (e *ChangesetDrainedEvent) event() {}

func (e *ChangesetDrainedEvent) Validate() error {
	return nil
}

//protobuf:7
type ChangesetFinalizedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetFinalizedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetFinalizedEvent{changeset: %s}", e.Key.String())
}

func (e *ChangesetFinalizedEvent) event() {}

func (e *ChangesetFinalizedEvent) Validate() error {
	return nil
}

//protobuf:8
type ChangesetRollingBackEvent struct {
	Key   key.Changeset `protobuf:"1"`
	Error string        `protobuf:"2"`
}

func (e *ChangesetRollingBackEvent) DebugString() string {
	return fmt.Sprintf("ChangesetRollingBackEvent{changeset: %s, Error: %s}", e.Key.String(), e.Error)
}

func (e *ChangesetRollingBackEvent) event() {}

func (e *ChangesetRollingBackEvent) Validate() error {
	return nil
}

//protobuf:9
type ChangesetFailedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetFailedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetFailedEvent{changeset: %s}", e.Key.String())
}

func (e *ChangesetFailedEvent) event() {}

func (e *ChangesetFailedEvent) Validate() error {
	return nil
}
