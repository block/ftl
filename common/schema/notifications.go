package schema

import (
	"fmt"

	"github.com/block/ftl/internal/key"
)

// TODO: these should be moved to the schema service package once go2proto supports referring to proto messages from other packages

//sumtype:decl
//protobuf:export
type Notification interface {
	notification()
}

// deployment Notifications
var _ Notification = (*FullSchemaNotification)(nil)
var _ Notification = (*DeploymentRuntimeNotification)(nil)

var _ Notification = (*ChangesetCreatedNotification)(nil)
var _ Notification = (*ChangesetPreparedNotification)(nil)
var _ Notification = (*ChangesetCommittedNotification)(nil)
var _ Notification = (*ChangesetFailedNotification)(nil)

//protobuf:1
type FullSchemaNotification struct {
	Schema     *Schema      `protobuf:"1"`
	Changesets []*Changeset `protobuf:"2"`
}

func (r *FullSchemaNotification) notification() {
}

//protobuf:2
type DeploymentRuntimeNotification struct {
	Payload   *RuntimeElement `protobuf:"1"`
	Changeset *key.Changeset  `protobuf:"2"`
}

func (e *DeploymentRuntimeNotification) notification() {}

//protobuf:3
type ChangesetCreatedNotification struct {
	Changeset *Changeset `protobuf:"1"`
}

func (e *ChangesetCreatedNotification) notification() {}

//protobuf:4
type ChangesetPreparedNotification struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetPreparedNotification) DebugString() string {
	return fmt.Sprintf("ChangesetPreparedNotification{changeset: %s}", e.Key.String())
}

func (e *ChangesetPreparedNotification) notification() {}

//protobuf:5
type ChangesetCommittedNotification struct {
	Changeset *Changeset `protobuf:"1"`
}

func (e *ChangesetCommittedNotification) notification() {}

//protobuf:6
type ChangesetDrainedNotification struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetDrainedNotification) notification() {}

//protobuf:7
type ChangesetFinalizedNotification struct {
	Key key.Changeset `protobuf:"1"`
}

func (e *ChangesetFinalizedNotification) notification() {}

//protobuf:8
type ChangesetRollingBackNotification struct {
	Changeset *Changeset `protobuf:"1"`
	Error     string     `protobuf:"2"`
}

func (e *ChangesetRollingBackNotification) notification() {}

//protobuf:9
type ChangesetFailedNotification struct {
	Key   key.Changeset `protobuf:"1"`
	Error string        `protobuf:"2"`
}

func (e *ChangesetFailedNotification) notification() {}
