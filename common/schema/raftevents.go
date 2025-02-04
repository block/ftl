package schema

import (
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

// TODO: these should be moved to the schema service package once go2proto supports referring to proto messages from other packages

//sumtype:decl
//protobuf:export
type Event interface {
	event()
	Validate() error
	// DebugString returns a string representation of the event for debugging purposes
	DebugString() string
}

// deployment events
var _ Event = (*DeploymentCreatedEvent)(nil)
var _ Event = (*DeploymentActivatedEvent)(nil)
var _ Event = (*DeploymentDeactivatedEvent)(nil)
var _ Event = (*DeploymentSchemaUpdatedEvent)(nil)
var _ Event = (*DeploymentReplicasUpdatedEvent)(nil)

// provisioner events
var _ Event = (*VerbRuntimeEvent)(nil)
var _ Event = (*TopicRuntimeEvent)(nil)
var _ Event = (*DatabaseRuntimeEvent)(nil)
var _ Event = (*ModuleRuntimeEvent)(nil)

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
type DeploymentSchemaUpdatedEvent struct {
	Key       key.Deployment `protobuf:"1"`
	Schema    *Module        `protobuf:"2"`
	Changeset key.Changeset  `protobuf:"3"`
}

func (r *DeploymentSchemaUpdatedEvent) DebugString() string {
	return fmt.Sprintf("DeploymentSchemaUpdatedEvent{key: %s, changeset: %s}", r.Key.String(), r.Changeset.String())
}

func (r *DeploymentSchemaUpdatedEvent) event() {}

func (r *DeploymentSchemaUpdatedEvent) Validate() error {
	return nil
}

//protobuf:3
type DeploymentReplicasUpdatedEvent struct {
	Key       key.Deployment `protobuf:"1"`
	Replicas  int            `protobuf:"2"`
	Changeset *key.Changeset `protobuf:"3"`
}

func (r *DeploymentReplicasUpdatedEvent) DebugString() string {
	return fmt.Sprintf("DeploymentReplicasUpdatedEvent{key: %s, changeset: %s, replicas: %d}", r.Key.String(), r.Changeset.String(), r.Replicas)
}

func (r *DeploymentReplicasUpdatedEvent) event() {}

func (r *DeploymentReplicasUpdatedEvent) Validate() error {
	return nil
}

//protobuf:4
type DeploymentActivatedEvent struct {
	Key         key.Deployment `protobuf:"1"`
	ActivatedAt time.Time      `protobuf:"2"`
	MinReplicas int            `protobuf:"3"`
	Changeset   *key.Changeset `protobuf:"4"`
}

func (r *DeploymentActivatedEvent) DebugString() string {
	return fmt.Sprintf("DeploymentActivatedEvent{key: %s, changeset: %s, minReplicas: %d}", r.Key.String(), r.Changeset.String(), r.MinReplicas)
}

func (r *DeploymentActivatedEvent) event() {}

func (r *DeploymentActivatedEvent) Validate() error {
	return nil
}

//protobuf:5
type DeploymentDeactivatedEvent struct {
	Key           key.Deployment `protobuf:"1"`
	ModuleRemoved bool           `protobuf:"2"`
	Changeset     *key.Changeset `protobuf:"3"`
}

func (r *DeploymentDeactivatedEvent) DebugString() string {
	return fmt.Sprintf("DeploymentDeactivatedEvent{key: %s, changeset: %s, moduleRemoved: %v}", r.Key.String(), r.Changeset.String(), r.ModuleRemoved)
}

func (r *DeploymentDeactivatedEvent) event() {}

func (r *DeploymentDeactivatedEvent) Validate() error {
	return nil
}

//protobuf:6
type VerbRuntimeEvent struct {
	Module       string                                   `protobuf:"1"`
	Changeset    key.Changeset                            `protobuf:"2"`
	ID           string                                   `protobuf:"3"`
	Subscription optional.Option[VerbRuntimeSubscription] `protobuf:"4"`
}

func (e *VerbRuntimeEvent) DebugString() string {
	return fmt.Sprintf("VerbRuntimeEvent{module: %s, changeset: %s, id: %v, subscription: %v}", e.Module, e.Changeset.String(), e.ID, e.Subscription)
}

func (e *VerbRuntimeEvent) event() {}

func (e *VerbRuntimeEvent) Validate() error {
	return nil
}

//protobuf:7
type TopicRuntimeEvent struct {
	Module    string        `protobuf:"1"`
	Changeset key.Changeset `protobuf:"2"`
	ID        string        `protobuf:"3"`
	Payload   *TopicRuntime `protobuf:"4"`
}

func (e *TopicRuntimeEvent) DebugString() string {
	return fmt.Sprintf("TopicRuntimeEvent{module: %s, changeset: %s, id: %v, payload: %v}", e.Module, e.Changeset.String(), e.ID, e.Payload)
}

func (e *TopicRuntimeEvent) event() {}

func (e *TopicRuntimeEvent) Validate() error {
	return nil
}

//protobuf:8
type DatabaseRuntimeEvent struct {
	Module      string                      `protobuf:"1"`
	Changeset   key.Changeset               `protobuf:"2"`
	ID          string                      `protobuf:"3"`
	Connections *DatabaseRuntimeConnections `protobuf:"4"`
}

func (e *DatabaseRuntimeEvent) DebugString() string {
	return fmt.Sprintf("DatabaseRuntimeEvent{module: %s, changeset: %s, id: %v, connections: %v}", e.Module, e.Changeset.String(), e.ID, e.Connections)
}

func (e *DatabaseRuntimeEvent) event() {}

func (e *DatabaseRuntimeEvent) Validate() error {
	return nil
}

//protobuf:9
type ModuleRuntimeEvent struct {
	DeploymentKey key.Deployment                           `protobuf:"1"`
	Changeset     *key.Changeset                           `protobuf:"2"`
	Base          optional.Option[ModuleRuntimeBase]       `protobuf:"3"`
	Scaling       optional.Option[ModuleRuntimeScaling]    `protobuf:"4"`
	Deployment    optional.Option[ModuleRuntimeDeployment] `protobuf:"5"`
}

func (e *ModuleRuntimeEvent) DebugString() string {
	return fmt.Sprintf("ModuleRuntimeEvent{deployment: %s, changeset: %s, deployment %v}", e.DeploymentKey.String(), e.Changeset.String(), e.Deployment)
}

func (e *ModuleRuntimeEvent) event() {}

func (e *ModuleRuntimeEvent) Validate() error {
	return nil
}

//protobuf:10
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

//protobuf:11
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

//protobuf:12
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

//protobuf:13
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

//protobuf:14
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

//protobuf:15
type ChangesetFailedEvent struct {
	Key   key.Changeset `protobuf:"1"`
	Error string        `protobuf:"2"`
}

func (e *ChangesetFailedEvent) DebugString() string {
	return fmt.Sprintf("ChangesetFailedEvent{changeset: %s, Error: %s}", e.Key.String(), e.Error)
}

func (e *ChangesetFailedEvent) event() {}

func (e *ChangesetFailedEvent) Validate() error {
	return nil
}
