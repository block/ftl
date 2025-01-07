package schemaeventsource

import (
	"context"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/rpc"
)

// View is a read-only view of the schema.
type View struct {
	view *atomic.Value[*schema.Schema]
}

// Get returns the current schema.
func (v *View) Get() *schema.Schema { return v.view.Load() }

// Event represents a change in the schema.
//
//sumtype:decl
type Event interface {
	// Schema is the READ-ONLY full schema after this event was applied.
	Schema() *schema.Schema
	change()
}

// EventRemove represents that a deployment (or module) was removed.
type EventRemove struct {
	// None for builtin modules.
	Deployment optional.Option[model.DeploymentKey]
	Module     *schema.Module

	schema *schema.Schema
}

func (c EventRemove) change()                {}
func (c EventRemove) Schema() *schema.Schema { return c.schema }

// EventUpsert represents that a module has been added or updated in the schema.
type EventUpsert struct {
	// schema is the READ-ONLY full schema after this event was applied.
	schema *schema.Schema

	// None for builtin modules.
	Deployment optional.Option[model.DeploymentKey]
	Module     *schema.Module
}

func (c EventUpsert) change()                {}
func (c EventUpsert) Schema() *schema.Schema { return c.schema }

// NewUnattached creates a new EventSource that is not attached to a SchemaService.
func NewUnattached() EventSource {
	return EventSource{
		events:              make(chan Event, 1024),
		view:                atomic.New[*schema.Schema](&schema.Schema{}),
		live:                atomic.New[bool](false),
		initialSyncComplete: make(chan struct{}),
	}
}

// EventSource represents a stream of schema events and the materialised view of those events.
type EventSource struct {
	events              chan Event
	view                *atomic.Value[*schema.Schema]
	live                *atomic.Value[bool]
	initialSyncComplete chan struct{}
}

// Events is a stream of schema change events.
//
// "View" will be updated with these changes prior to being sent on this channel.
//
// NOTE: Only a single goroutine should read from the EventSource.
func (e EventSource) Events() <-chan Event { return e.events }

// ViewOnly converts the EventSource into a read-only view of the schema.
//
// This will consume all events so the EventSource dodesn't block as the view is automatically updated.
func (e EventSource) ViewOnly() View {
	go func() {
		for range e.Events() { //nolint:revive
		}
	}()
	return View{e.view}
}

// Live returns true if the EventSource is connected to the SchemaService.
func (e EventSource) Live() bool { return e.live.Load() }

// WaitForInitialSync blocks until the initial sync has completed or the context is cancelled.
//
// Returns true if the initial sync has completed, false if the context was cancelled.
func (e EventSource) WaitForInitialSync(ctx context.Context) bool {
	select {
	case <-e.initialSyncComplete:
		return true

	case <-ctx.Done():
		return false
	}
}

// View is the materialised view of the schema from "Events".
func (e EventSource) View() *schema.Schema { return e.view.Load() }

// Publish an event to the EventSource.
//
// This will update the materialised view and send the event on the "Events" channel. The event will be updated with the
// materialised view.
//
// This is mostly useful in conjunction with NewUnattached, for testing.
func (e EventSource) Publish(event Event) {
	clone := reflect.DeepCopy(e.View())
	switch event := event.(type) {
	case EventRemove:
		clone.Modules = slices.DeleteFunc(clone.Modules, func(m *schema.Module) bool { return m.Name == event.Module.Name })
		event.schema = clone
		e.view.Store(clone)
		e.events <- event

	case EventUpsert:
		if i := slices.IndexFunc(clone.Modules, func(m *schema.Module) bool { return m.Name == event.Module.Name }); i != -1 {
			clone.Modules[i] = event.Module
		} else {
			clone.Modules = append(clone.Modules, event.Module)
		}
		event.schema = clone
		e.view.Store(clone)
		e.events <- event
	}
}

// New creates a new EventSource that pulls schema changes from the SchemaService into an event channel and a
// materialised view (ie. [schema.Schema]).
//
// The sync will terminate when the context is cancelled.
func New(ctx context.Context, client ftlv1connect.SchemaServiceClient) EventSource {
	logger := log.FromContext(ctx).Scope("schema-sync")
	out := NewUnattached()
	initialSyncComplete := false
	logger.Debugf("Starting schema pull")

	// Set the initial "live" state by pinging the server. After that we'll rely on the stream.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := client.Ping(pingCtx, connect.NewRequest(&ftlv1.PingRequest{}))
	out.live.Store(err == nil && resp.Msg.NotReady == nil)

	go rpc.RetryStreamingServerStream(ctx, "schema-sync", backoff.Backoff{}, &ftlv1.WatchRequest{}, client.Watch, func(_ context.Context, resp *ftlv1.WatchResponse) error {
		out.live.Store(true)
		modules := make([]*schema.Module, len(resp.Schema.Modules))
		for i, module := range resp.Schema.Modules {
			m, err := schema.ModuleFromProto(module)
			if err != nil {
				return fmt.Errorf("schema-sync: failed to decode module schema: %w", err)
			}
			modules[i] = m
		}

		// do not report changes until the initial sync is complete
		if initialSyncComplete {
			currentModules := out.View().Modules
			events := changeEvents(ctx, currentModules, modules)
			for _, event := range events {
				out.Publish(event)
			}
		}

		out.view.Store(&schema.Schema{Modules: modules})

		if !initialSyncComplete {
			initialSyncComplete = true
			out.initialSyncComplete <- struct{}{}
			close(out.initialSyncComplete)
		}

		return nil
	}, func(_ error) bool {
		out.live.Store(false)
		return true
	})
	return out
}

func changeEvents(ctx context.Context, currentModules []*schema.Module, newModules []*schema.Module) []Event {
	logger := log.FromContext(ctx).Scope("schema-sync")

	currentByName := map[string]*schema.Module{}
	for _, module := range currentModules {
		currentByName[module.Name] = module
	}
	newByName := map[string]*schema.Module{}
	for _, module := range newModules {
		newByName[module.Name] = module
	}

	result := []Event{}
	for _, sch := range newModules {
		if current, ok := currentByName[sch.Name]; !ok {
			logger.Tracef("Module %s added", sch.Name)
			result = append(result, EventUpsert{
				Deployment: deploymentKey(sch),
				Module:     sch,
			})
		} else if !current.Equals(sch) {
			logger.Tracef("Module %s updated", sch.Name)
			result = append(result, EventUpsert{
				Deployment: deploymentKey(sch),
				Module:     sch,
			})
		}
	}

	for _, module := range currentModules {
		if _, ok := newByName[module.Name]; !ok {
			logger.Tracef("Module %s removed", module.Name)
			result = append(result, EventRemove{
				Deployment: deploymentKey(module),
				Module:     module,
			})
		}
	}

	return result
}

// TODO: can this be removed from the events?
func deploymentKey(module *schema.Module) optional.Option[model.DeploymentKey] {
	if module.Runtime == nil || module.Runtime.Deployment == nil || module.Runtime.Deployment.DeploymentKey == "" {
		return optional.None[model.DeploymentKey]()
	}
	deploymentKey, err := model.ParseDeploymentKey(module.Runtime.Deployment.DeploymentKey)
	if err != nil {
		return optional.None[model.DeploymentKey]()
	}
	return optional.Some(deploymentKey)
}
