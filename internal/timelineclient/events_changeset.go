package timelineclient

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

// ChangesetCreated represents a timeline event for when a changeset is created
type ChangesetCreated struct {
	Key       key.Changeset
	CreatedAt time.Time
	Modules   []string // Names of modules being added or modified
	ToRemove  []string // Names of modules being removed
}

var _ Event = ChangesetCreated{}

func (ChangesetCreated) clientEvent() {}
func (c ChangesetCreated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_ChangesetCreated{
			ChangesetCreated: &timelinepb.ChangesetCreatedEvent{
				Key:       c.Key.String(),
				CreatedAt: timestamppb.New(c.CreatedAt),
				Modules:   c.Modules,
				ToRemove:  c.ToRemove,
			},
		},
	}, nil
}

// DeploymentCreated represents a timeline event for when a deployment is created
type DeploymentCreated struct {
	Key       key.Deployment
	CreatedAt time.Time
	Schema    *schema.Module
	Changeset optional.Option[key.Changeset]
}

var _ Event = DeploymentCreated{}

func (DeploymentCreated) clientEvent() {}
func (d DeploymentCreated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	// Create a new DeploymentCreatedEvent for the timeline
	event := &timelinepb.DeploymentCreatedEvent{
		Key:       d.Key.String(),
		CreatedAt: timestamppb.New(d.CreatedAt),
		Module:    d.Schema.Name,
	}

	// Add changeset information if available
	if val, ok := d.Changeset.Get(); ok {
		str := val.String()
		event.Changeset = &str
	}

	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated{
			DeploymentCreated: event,
		},
	}, nil
}

// DeploymentRuntime represents a timeline event for deployment runtime changes
type DeploymentRuntime struct {
	Deployment key.Deployment
	Changeset  optional.Option[key.Changeset]
	Element    *schema.RuntimeElement
	UpdatedAt  time.Time
}

var _ Event = DeploymentRuntime{}

func (DeploymentRuntime) clientEvent() {}
func (d DeploymentRuntime) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	// Create a new DeploymentRuntimeEvent for the timeline
	event := &timelinepb.DeploymentRuntimeEvent{
		Key:       d.Deployment.String(),
		UpdatedAt: timestamppb.New(d.UpdatedAt),
	}

	// Add element name if available
	if val, ok := d.Element.Name.Get(); ok {
		event.ElementName = &val
	}

	// Add element type based on the runtime element
	if d.Element != nil && d.Element.Element != nil {
		switch d.Element.Element.(type) {
		case *schema.ModuleRuntimeDeployment:
			event.ElementType = "deployment"
		case *schema.ModuleRuntimeScaling:
			event.ElementType = "scaling"
		case *schema.ModuleRuntimeRunner:
			event.ElementType = "runner"
		case *schema.VerbRuntime:
			event.ElementType = "verb"
		case *schema.TopicRuntime:
			event.ElementType = "topic"
		case *schema.DatabaseRuntime:
			event.ElementType = "database"
		default:
			event.ElementType = "unknown"
		}
	}

	// Add changeset information if available
	if val, ok := d.Changeset.Get(); ok {
		str := val.String()
		event.Changeset = &str
	}

	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentRuntime{
			DeploymentRuntime: event,
		},
	}, nil
}

// ChangesetStateChanged represents a timeline event for when a changeset changes state
type ChangesetStateChanged struct {
	Key   key.Changeset
	State schema.ChangesetState
	Error optional.Option[string] // Present if state is FAILED
}

var _ Event = ChangesetStateChanged{}

func (ChangesetStateChanged) clientEvent() {}
func (c ChangesetStateChanged) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_ChangesetStateChanged{
			ChangesetStateChanged: &timelinepb.ChangesetStateChangedEvent{
				Key:   c.Key.String(),
				State: c.State.ToProto(),
				Error: c.Error.Ptr(),
			},
		},
	}, nil
}
