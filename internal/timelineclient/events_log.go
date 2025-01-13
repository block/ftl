package timelineclient

import (
	"time"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/internal/key"
)

type Log struct {
	DeploymentKey key.Deployment
	RequestKey    optional.Option[key.Request]
	Time          time.Time
	Level         int32
	Attributes    map[string]string
	Message       string
	Error         optional.Option[string]
}

var _ Event = Log{}

func (Log) clientEvent() {}
func (l Log) ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error) {
	var requestKey *string
	if r, ok := l.RequestKey.Get(); ok {
		key := r.String()
		requestKey = &key
	}
	return &timelinepb.CreateEventsRequest_EventEntry{
		Entry: &timelinepb.CreateEventsRequest_EventEntry_Log{
			Log: &timelinepb.LogEvent{
				DeploymentKey: l.DeploymentKey.String(),
				RequestKey:    requestKey,
				Timestamp:     timestamppb.New(l.Time),
				LogLevel:      l.Level,
				Attributes:    l.Attributes,
				Message:       l.Message,
				Error:         l.Error.Ptr(),
			},
		},
	}, nil
}
