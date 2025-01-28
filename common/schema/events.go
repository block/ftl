package schema

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/key"
)

// TODO: these should be moved to the schema service package once go2proto supports referring to proto messages from other packages

//sumtype:decl
//protobuf:export
type Event interface {
	event()
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

//protobuf:1
type DeploymentCreatedEvent struct {
	Key    key.Deployment `protobuf:"1"`
	Schema *Module        `protobuf:"2"`
}

func (r *DeploymentCreatedEvent) event() {}

//protobuf:2
type DeploymentSchemaUpdatedEvent struct {
	Key       key.Deployment `protobuf:"1"`
	Schema    *Module        `protobuf:"2"`
	Changeset *key.Changeset `protobuf:"3"`
}

func (r *DeploymentSchemaUpdatedEvent) event() {}

//protobuf:3
type DeploymentReplicasUpdatedEvent struct {
	Key       key.Deployment `protobuf:"1"`
	Replicas  int            `protobuf:"2"`
	Changeset *key.Changeset `protobuf:"3"`
}

func (r *DeploymentReplicasUpdatedEvent) event() {}

//protobuf:4
type DeploymentActivatedEvent struct {
	Key         key.Deployment `protobuf:"1"`
	ActivatedAt time.Time      `protobuf:"2"`
	MinReplicas int            `protobuf:"3"`
	Changeset   *key.Changeset `protobuf:"4"`
}

func (r *DeploymentActivatedEvent) event() {}

//protobuf:5
type DeploymentDeactivatedEvent struct {
	Key           key.Deployment `protobuf:"1"`
	ModuleRemoved bool           `protobuf:"2"`
	Changeset     *key.Changeset `protobuf:"3"`
}

func (r *DeploymentDeactivatedEvent) event() {}

//protobuf:6
type VerbRuntimeEvent struct {
	Module       string                                   `protobuf:"1"`
	ID           string                                   `protobuf:"2"`
	Base         optional.Option[VerbRuntimeBase]         `protobuf:"3"`
	Subscription optional.Option[VerbRuntimeSubscription] `protobuf:"4"`
}

func (e *VerbRuntimeEvent) event() {}

//protobuf:7
type TopicRuntimeEvent struct {
	Module  string        `protobuf:"1"`
	ID      string        `protobuf:"2"`
	Payload *TopicRuntime `protobuf:"3"`
}

func (e *TopicRuntimeEvent) event() {}

//protobuf:8
type DatabaseRuntimeEvent struct {
	Module      string                      `protobuf:"1"`
	ID          string                      `protobuf:"2"`
	Connections *DatabaseRuntimeConnections `protobuf:"3"`
}

func (e *DatabaseRuntimeEvent) event() {}

//protobuf:9
type ModuleRuntimeEvent struct {
	Module string `protobuf:"1"`
	// None if updating at provisioning
	DeploymentKey optional.Option[string] `protobuf:"2"`

	Base       optional.Option[ModuleRuntimeBase]       `protobuf:"3"`
	Scaling    optional.Option[ModuleRuntimeScaling]    `protobuf:"4"`
	Deployment optional.Option[ModuleRuntimeDeployment] `protobuf:"5"`
}

func (e *ModuleRuntimeEvent) event() {}

//protobuf:10
type ProvisioningCreatedEvent struct {
	DesiredModule *Module `protobuf:"1"`
}

func (e *ProvisioningCreatedEvent) event() {}

//protobuf:11
type ChangesetCreatedEvent struct {
	Changeset *Changeset `protobuf:"1"`
}

func (e *ChangesetCreatedEvent) event() {}

//protobuf:12
type ChangesetCommittedEvent struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetCommittedEvent) event() {}

//protobuf:13
type ChangesetFailedEvent struct {
	Key   key.Changeset `protobuf:"1"`
	Error string        `protobuf:"2"`
}

func (e *ChangesetFailedEvent) event() {}
