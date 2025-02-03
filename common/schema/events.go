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

func (r *DeploymentDeactivatedEvent) event() {}

func (r *DeploymentDeactivatedEvent) Validate() error {
	return nil
}

//protobuf:6
type VerbRuntimeEvent struct {
	Module       string                                   `protobuf:"1"`
	Changeset    key.Changeset                            `protobuf:"2"`
	ID           string                                   `protobuf:"3"`
	Base         optional.Option[VerbRuntimeBase]         `protobuf:"4"`
	Subscription optional.Option[VerbRuntimeSubscription] `protobuf:"5"`
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

func (e *ModuleRuntimeEvent) event() {}

func (e *ModuleRuntimeEvent) Validate() error {
	return nil
}

//protobuf:11
type ChangesetCreatedEvent struct {
	Changeset *Changeset `protobuf:"1"`
}

func (e *ChangesetCreatedEvent) event() {}

func (e *ChangesetCreatedEvent) Validate() error {
	if e.Changeset.Key.IsZero() {
		return fmt.Errorf("changeset key is required")
	}
	return nil
}

//protobuf:12
type ChangesetCommittedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetCommittedEvent) event() {}

func (e *ChangesetCommittedEvent) Validate() error {
	return nil
}

//protobuf:13
type ChangesetPreparedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetPreparedEvent) event() {}

func (e *ChangesetPreparedEvent) Validate() error {
	return nil
}

//protobuf:14
type ChangesetFailedEvent struct {
	Key   key.Changeset `protobuf:"1"`
	Error string        `protobuf:"2"`
}

func (e *ChangesetFailedEvent) event() {}

func (e *ChangesetFailedEvent) Validate() error {
	return nil
}
