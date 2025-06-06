package timelineclient

import (
	"context"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/must"
	"google.golang.org/protobuf/types/known/timestamppb"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1/timelinepbconnect"
	v1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/rpc"
)

const (
	maxBatchSize  = 16
	maxBatchDelay = 100 * time.Millisecond
)

type Config struct {
	TimelineEndpoint *url.URL `help:"Timeline endpoint (discard:// to disable)." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
}

type Publisher interface {
	Publish(ctx context.Context, event Event)
}

type RealClient struct {
	timelinepbconnect.TimelineServiceClient

	entries          chan *timelinepb.CreateEventsRequest_EventEntry
	lastDroppedError atomic.Value[time.Time]
	lastFailedError  atomic.Value[time.Time]
}

var _ Publisher = &RealClient{}

func (c *RealClient) Ping(context.Context, *connect.Request[v1.PingRequest]) (*connect.Response[v1.PingResponse], error) {
	return connect.NewResponse(&v1.PingResponse{}), nil
}

var NullConfig = Config{
	TimelineEndpoint: must.Get(url.Parse("discard://")),
}

// NewClient creates a new Timeline client.
//
// If endpoint is discard:// the client will not create an RPC client or send any RPC requests, and all events
// will be immediately discarded.
func NewClient(ctx context.Context, config Config) *RealClient {
	var c timelinepbconnect.TimelineServiceClient
	if config.TimelineEndpoint.Scheme != "discard" {
		c = rpc.Dial(timelinepbconnect.NewTimelineServiceClient, config.TimelineEndpoint.String(), log.Error)
	}
	client := &RealClient{
		TimelineServiceClient: c,
		entries:               make(chan *timelinepb.CreateEventsRequest_EventEntry, 1000),
	}
	if config.TimelineEndpoint.Scheme == "discard" {
		go client.noopEvents(ctx)
	} else {
		go client.processEvents(ctx)
	}
	return client
}

//go:sumtype
type Event interface {
	ToEntry() (*timelinepb.CreateEventsRequest_EventEntry, error)
	clientEvent()
}

// Publish asynchronously enqueues an event for publication to the timeline.
func (c *RealClient) Publish(ctx context.Context, event Event) {
	entry, err := event.ToEntry()
	entry.Timestamp = timestamppb.New(time.Now())
	if err != nil {
		log.FromContext(ctx).Warnf("failed to create request to publish %T event: %v", event, err)
		return
	}
	select {
	case c.entries <- entry:
	default:
		if time.Since(c.lastDroppedError.Load()) > 10*time.Second {
			log.FromContext(ctx).Warnf("Dropping event %T due to full queue", event)
			c.lastDroppedError.Store(time.Now())
		}
	}
}

func (c *RealClient) noopEvents(ctx context.Context) {
	for {
		select {
		case <-c.entries:
		case <-ctx.Done():
			return
		}
	}
}

func (c *RealClient) processEvents(ctx context.Context) {
	lastFlush := time.Now()
	buffer := make([]*timelinepb.CreateEventsRequest_EventEntry, 0, maxBatchSize)
	for {
		select {
		case <-ctx.Done():
			return

		case entry := <-c.entries:
			buffer = append(buffer, entry)

			if len(buffer) < maxBatchSize || time.Since(lastFlush) < maxBatchDelay {
				continue
			}
			c.flushEvents(ctx, buffer)
			buffer = nil

		case <-time.After(maxBatchDelay):
			if len(buffer) == 0 {
				continue
			}
			c.flushEvents(ctx, buffer)
			buffer = nil
		}
	}
}

// Flush all events in the buffer to the timeline service in a single call.
func (c *RealClient) flushEvents(ctx context.Context, entries []*timelinepb.CreateEventsRequest_EventEntry) {
	logger := log.FromContext(ctx).Scope("timeline")
	_, err := c.CreateEvents(ctx, connect.NewRequest(&timelinepb.CreateEventsRequest{
		Entries: entries,
	}))
	if err != nil {
		if time.Since(c.lastFailedError.Load()) > 10*time.Second {
			logger.Errorf(err, "Failed to insert %d events", len(entries))
			c.lastFailedError.Store(time.Now())
		}
		metrics.Failed(ctx, len(entries))
		return
	}
	metrics.Inserted(ctx, len(entries))
}

// NewFakePublisher for testing.
func NewFakePublisher() *FakePublisher {
	return &FakePublisher{}
}

type FakePublisher struct {
	Events []Event
}

func (f *FakePublisher) Publish(_ context.Context, event Event) {
	f.Events = append(f.Events, event)
}
