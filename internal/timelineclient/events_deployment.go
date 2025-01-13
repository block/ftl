package timelineclient

import (
	"time"

	"github.com/alecthomas/types/optional"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/internal/key"
)

type DeploymentCreated struct {
	DeploymentKey      key.Deployment
	Time               time.Time
	ModuleName         string
	MinReplicas        int
	ReplacedDeployment optional.Option[key.Deployment]
}

var _ Event = DeploymentCreated{}

func (DeploymentCreated) clientEvent() {}
func (d DeploymentCreated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	var replaced *string
	if r, ok := d.ReplacedDeployment.Get(); ok {
		repl := r.String()
		replaced = &repl
	}
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentCreated{
			DeploymentCreated: &timelinepb.DeploymentCreatedEvent{
				Key:         d.DeploymentKey.String(),
				ModuleName:  d.ModuleName,
				MinReplicas: int32(d.MinReplicas),
				Replaced:    replaced,
			},
		},
	}, nil
}

type DeploymentUpdated struct {
	DeploymentKey   key.Deployment
	Time            time.Time
	MinReplicas     int
	PrevMinReplicas int
}

var _ Event = DeploymentUpdated{}

func (DeploymentUpdated) clientEvent() {}
func (d DeploymentUpdated) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_DeploymentUpdated{
			DeploymentUpdated: &timelinepb.DeploymentUpdatedEvent{
				Key:             d.DeploymentKey.String(),
				MinReplicas:     int32(d.MinReplicas),
				PrevMinReplicas: int32(d.PrevMinReplicas),
			},
		},
	}, nil
}
