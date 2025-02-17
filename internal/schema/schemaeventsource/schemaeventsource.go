package schemaeventsource

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/pubsub"
	"github.com/jpillora/backoff"
	"golang.org/x/exp/maps"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

type PullSchemaClient interface {
	PullSchema(context.Context, *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
	Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

// View is a read-only view of the schema.
type View struct {
	eventSource *EventSource
}

type currentState struct {
	schema           *schema.Schema
	activeChangesets map[key.Changeset]*schema.Changeset
}

// GetCanonical returns the current canonical schema (ie: without any changes applied from active changesets)
func (v *View) GetCanonical() *schema.Schema { return v.eventSource.view.Load().schema }

// NewUnattached creates a new EventSource that is not attached to a SchemaService.
func NewUnattached() *EventSource {
	return &EventSource{
		events:              pubsub.New[schema.Notification](),
		view:                atomic.New(&currentState{schema: &schema.Schema{}, activeChangesets: map[key.Changeset]*schema.Changeset{}}),
		live:                atomic.New[bool](false),
		initialSyncComplete: make(chan struct{}),
		subscribeLock:       &sync.Mutex{},
	}
}

// EventSource represents a stream of schema events and the materialised view of those events.
type EventSource struct {
	events              *pubsub.Topic[schema.Notification]
	view                *atomic.Value[*currentState]
	live                *atomic.Value[bool]
	subscribeLock       *sync.Mutex
	initialSyncComplete chan struct{}
	initialSync         bool
}

// Subscribe subscribes you to the schema events
//
// This method guarentes you will always receive a FullSchemaNotification as the first message
func (e *EventSource) Subscribe(ctx context.Context) <-chan schema.Notification {
	e.subscribeLock.Lock()
	defer e.subscribeLock.Unlock()
	subscribe := e.events.Subscribe(nil)
	context.AfterFunc(ctx, func() {
		e.events.Unsubscribe(subscribe)
	})
	// We always send a full schema event
	select {
	case <-e.initialSyncComplete:
		// Initial sync is complete, we send an initial Full schema event
		state := e.view.Load()
		subscribe <- &schema.FullSchemaNotification{Schema: state.schema, Changesets: maps.Values(state.activeChangesets)}
	default:

	}
	return subscribe
}

// ViewOnly converts the EventSource into a read-only view of the schema.
func (e *EventSource) ViewOnly() *View {
	return &View{eventSource: e}

}

// Live returns true if the EventSource is connected to the SchemaService.
func (e *EventSource) Live() bool { return e.live.Load() }

// WaitForInitialSync blocks until the initial sync has completed or the context is cancelled.
//
// Returns true if the initial sync has completed, false if the context was cancelled.
func (e *EventSource) WaitForInitialSync(ctx context.Context) bool {
	select {
	case <-e.initialSyncComplete:
		return true

	case <-ctx.Done():
		return false
	}
}

// CanonicalView is the materialised view of the schema from "Events".
func (e *EventSource) CanonicalView() *schema.Schema { return e.view.Load().schema }

func (e *EventSource) ActiveChangeset() map[key.Changeset]*schema.Changeset {
	return e.view.Load().activeChangesets
}

func (e *EventSource) PublishModuleForTest(module *schema.Module) error {
	return e.Publish(&schema.FullSchemaNotification{Schema: &schema.Schema{Modules: []*schema.Module{module}}})
}

// Publish an event to the EventSource.
//
// This will update the materialised view and send the event on the "Events" channel. The event will be updated with the
// materialised view.
//
// This is mostly useful in conjunction with NewUnattached, for testing.
func (e *EventSource) Publish(event schema.Notification) error {
	e.subscribeLock.Lock()
	defer e.subscribeLock.Unlock()
	switch event := event.(type) {
	case *schema.FullSchemaNotification:
		changesets := map[key.Changeset]*schema.Changeset{}
		for _, cs := range event.Changesets {
			changesets[cs.Key] = cs
		}
		e.view.Store(&currentState{schema: event.Schema, activeChangesets: changesets})
		if !e.initialSync {
			e.initialSync = true
			close(e.initialSyncComplete)
		}
	case *schema.DeploymentRuntimeNotification:
		clone := reflect.DeepCopy(e.view.Load())
		if event.Changeset != nil && !event.Changeset.IsZero() {
			cs := clone.activeChangesets[*event.Changeset]
			for _, m := range cs.Modules {
				if m.Runtime.Deployment.DeploymentKey == event.Payload.Deployment {
					err := event.Payload.ApplyToModule(m)
					if err != nil {
						return fmt.Errorf("failed to apply deployment runtime: %w", err)
					}
					break
				}
			}
		} else {
			for _, m := range clone.schema.Modules {
				if m.Runtime.Deployment.DeploymentKey == event.Payload.Deployment {
					err := event.Payload.ApplyToModule(m)
					if err != nil {
						return fmt.Errorf("failed to apply deployment runtime: %w", err)
					}
					break
				}
			}
		}
		e.view.Store(clone)
	case *schema.ChangesetCreatedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		clone.activeChangesets[event.Changeset.Key] = event.Changeset
		e.view.Store(clone)
	case *schema.ChangesetPreparedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		cs := clone.activeChangesets[event.Key]
		for _, module := range cs.Modules {
			module.Runtime.Deployment.State = schema.DeploymentStateCanary
		}
		e.view.Store(clone)
	case *schema.ChangesetCommittedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		clone.activeChangesets[event.Changeset.Key] = event.Changeset
		modules := clone.schema.Modules
		for _, module := range event.Changeset.Modules {
			module.Runtime.Deployment.State = schema.DeploymentStateCanonical
			if i := slices.IndexFunc(modules, func(m *schema.Module) bool { return m.Name == module.Name }); i != -1 {
				modules[i] = module
			} else {
				modules = append(modules, module)
			}
		}
		for _, removed := range event.Changeset.RemovingModules {
			modules = islices.Filter(modules, func(m *schema.Module) bool {
				return m.ModRuntime().ModDeployment().DeploymentKey != removed.ModRuntime().ModDeployment().DeploymentKey
			})

		}
		clone.schema.Modules = modules
		e.view.Store(clone)
	case *schema.ChangesetDrainedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		cs := clone.activeChangesets[event.Key]
		for _, module := range cs.OwnedModules() {
			module.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
		}
		e.view.Store(clone)
	case *schema.ChangesetRollingBackNotification:
		clone := reflect.DeepCopy(e.view.Load())
		cs := event.Changeset
		for _, module := range cs.Modules {
			module.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
		}
		cs.Error = event.Error
		e.view.Store(clone)
	case *schema.ChangesetFailedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		delete(clone.activeChangesets, event.Key)
		e.view.Store(clone)
	case *schema.ChangesetFinalizedNotification:
		clone := reflect.DeepCopy(e.view.Load())
		delete(clone.activeChangesets, event.Key)
		e.view.Store(clone)
	}
	e.events.Publish(event)
	return nil
}

// New creates a new EventSource that pulls schema changes from the SchemaService into an event channel and a
// materialised view (ie. [schema.Schema]).
//
// The sync will terminate when the context is cancelled.
func New(ctx context.Context, subscriptionID string, client PullSchemaClient) *EventSource {
	logger := log.FromContext(ctx).Scope("schema-sync")
	out := NewUnattached()
	logger.Debugf("Starting schema pull")

	// Set the initial "live" state by pinging the server. After that we'll rely on the stream.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := client.Ping(pingCtx, connect.NewRequest(&ftlv1.PingRequest{}))
	out.live.Store(err == nil && resp.Msg.NotReady == nil)

	logger.Tracef("Schema pull live: %t", out.live.Load())

	go rpc.RetryStreamingServerStream(ctx, "schema-sync", backoff.Backoff{}, &ftlv1.PullSchemaRequest{SubscriptionId: subscriptionID}, client.PullSchema, func(_ context.Context, resp *ftlv1.PullSchemaResponse) error {
		out.live.Store(true)

		logger.Tracef("Schema pull %s (event: %T)", subscriptionID, resp.Event.Value)

		proto, err := schema.NotificationFromProto(resp.Event)
		if err != nil {
			return fmt.Errorf("failed to decode schema event: %w", err)
		}
		err = out.Publish(proto)
		if err != nil {
			logger.Errorf(err, "Failed to publish schema event")
			return fmt.Errorf("failed to publish schema event: %w", err)
		}
		return nil
	}, func(_ error) bool {
		out.live.Store(false)
		return true
	})
	return out
}
