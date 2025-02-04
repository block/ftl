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
	islices "github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

// View is a read-only view of the schema.
type View struct {
	view            *atomic.Value[*schema.Schema]
	activeChangeset *atomic.Value[optional.Option[*schema.Changeset]]
}

// GetCanonical returns the current canonical schema (ie: without any changes applied from active changesets)
func (v *View) GetCanonical() *schema.Schema { return v.view.Load() }

// GetLatest returns the latest schema, by applying active deployments in changesetz to the canonical schema.
func (v *View) GetLatest() *schema.Schema {
	return latestSchema(v.GetCanonical(), v.GetActiveChangeset())
}

// latest schema calculates the latest schema by applying active deployments in changeset to the canonical schema.
func latestSchema(canonical *schema.Schema, changeset optional.Option[*schema.Changeset]) *schema.Schema {
	ch, ok := changeset.Get()
	if !ok {
		return canonical
	}
	sch := reflect.DeepCopy(canonical)
	for _, module := range ch.Modules {
		if i := slices.IndexFunc(sch.Modules, func(m *schema.Module) bool { return m.Name == module.Name }); i != -1 {
			sch.Modules[i] = module
		} else {
			sch.Modules = append(sch.Modules, module)
		}
	}
	return sch
}

// GetActiveChangeset returns the current active changeset.
func (v *View) GetActiveChangeset() optional.Option[*schema.Changeset] {
	return v.activeChangeset.Load()
}

// Event represents a change in the schema.
//
//sumtype:decl
type Event interface {
	// More returns true if there are more changes to come as part of the initial sync.
	More() bool
	// GetCanonical is the READ-ONLY canonical schema after this event was applied.
	GetCanonical() *schema.Schema
	// GetLatest is the READ-ONLY latest schema, by applying active deployments in changesetz to the canonical schema.
	GetLatest() *schema.Schema
	// ActiveChangeset is the READ-ONLY active changeset after this event was applied.
	ActiveChangeset() optional.Option[*schema.Changeset]
	finalize(canonical *schema.Schema, activeChangeset optional.Option[*schema.Changeset]) Event
	change()
}

// EventRemove represents that a deployment (or module) was removed from the active canonical schema.
// This is different than when a deployment is replaced by a changeset successfully being committed. See EventChangesetEnded
type EventRemove struct {
	// None for builtin modules.
	Deployment optional.Option[key.Deployment]
	Module     string

	schema          *schema.Schema
	activeChangeset optional.Option[*schema.Changeset]
	more            bool
}

func (c EventRemove) change()                      {}
func (c EventRemove) More() bool                   { return c.more }
func (c EventRemove) GetCanonical() *schema.Schema { return c.schema }
func (c EventRemove) GetLatest() *schema.Schema {
	return latestSchema(c.schema, c.activeChangeset)
}
func (c EventRemove) ActiveChangeset() optional.Option[*schema.Changeset] { return c.activeChangeset }
func (c EventRemove) finalize(sch *schema.Schema, activeChangeset optional.Option[*schema.Changeset]) Event {
	c.schema = sch
	c.activeChangeset = activeChangeset
	return c
}

// EventUpsert represents that a module has been added or updated in the schema.
type EventUpsert struct {
	// None for builtin modules.
	Module    *schema.Module
	Changeset optional.Option[key.Changeset]

	schema          *schema.Schema
	activeChangeset optional.Option[*schema.Changeset]
	more            bool
}

func (c EventUpsert) change()                      {}
func (c EventUpsert) More() bool                   { return c.more }
func (c EventUpsert) GetCanonical() *schema.Schema { return c.schema }
func (c EventUpsert) GetLatest() *schema.Schema {
	return latestSchema(c.schema, c.activeChangeset)
}
func (c EventUpsert) ActiveChangeset() optional.Option[*schema.Changeset] { return c.activeChangeset }
func (c EventUpsert) finalize(sch *schema.Schema, activeChangeset optional.Option[*schema.Changeset]) Event {
	c.schema = sch
	c.activeChangeset = activeChangeset
	return c
}

type EventChangesetStarted struct {
	Changeset *schema.Changeset

	schema          *schema.Schema
	activeChangeset optional.Option[*schema.Changeset]
	more            bool
}

func (c EventChangesetStarted) change()                      {}
func (c EventChangesetStarted) More() bool                   { return c.more }
func (c EventChangesetStarted) GetCanonical() *schema.Schema { return c.schema }
func (c EventChangesetStarted) GetLatest() *schema.Schema {
	return latestSchema(c.schema, c.activeChangeset)
}
func (c EventChangesetStarted) ActiveChangeset() optional.Option[*schema.Changeset] {
	return c.activeChangeset
}
func (c EventChangesetStarted) finalize(sch *schema.Schema, activeChangeset optional.Option[*schema.Changeset]) Event {
	c.schema = sch
	c.activeChangeset = activeChangeset
	return c
}

type EventChangesetEnded struct {
	Key     key.Changeset
	Success bool
	Error   string

	// ReplacedDeloyments contains each deployment that was superseeded by this changeset
	// If Success is false it is empty
	ReplacedDeloyments []key.Deployment

	schema          *schema.Schema
	activeChangeset optional.Option[*schema.Changeset]
	more            bool
}

func (c EventChangesetEnded) change()                      {}
func (c EventChangesetEnded) More() bool                   { return c.more }
func (c EventChangesetEnded) GetCanonical() *schema.Schema { return c.schema }
func (c EventChangesetEnded) GetLatest() *schema.Schema {
	return latestSchema(c.schema, c.activeChangeset)
}
func (c EventChangesetEnded) ActiveChangeset() optional.Option[*schema.Changeset] {
	return c.activeChangeset
}
func (c EventChangesetEnded) finalize(sch *schema.Schema, activeChangeset optional.Option[*schema.Changeset]) Event {
	c.schema = sch
	c.activeChangeset = activeChangeset
	return c
}

// NewUnattached creates a new EventSource that is not attached to a SchemaService.
func NewUnattached() EventSource {
	return EventSource{
		events:              make(chan Event, 1024),
		view:                atomic.New(&schema.Schema{}),
		activeChangeset:     atomic.New(optional.None[*schema.Changeset]()),
		live:                atomic.New[bool](false),
		initialSyncComplete: make(chan struct{}),
	}
}

// EventSource represents a stream of schema events and the materialised view of those events.
type EventSource struct {
	events              chan Event
	view                *atomic.Value[*schema.Schema]
	activeChangeset     *atomic.Value[optional.Option[*schema.Changeset]]
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
	return View{
		view:            e.view,
		activeChangeset: e.activeChangeset,
	}
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

// CanonicalView is the materialised view of the schema from "Events".
func (e EventSource) CanonicalView() *schema.Schema { return e.view.Load() }

// LatestView is the materialised view of the schema from "Events" taking into account active deployments in changesets.
func (e EventSource) LatestView() *schema.Schema {
	return latestSchema(e.view.Load(), e.activeChangeset.Load())
}

func (e EventSource) ActiveChangeset() optional.Option[*schema.Changeset] {
	return e.activeChangeset.Load()
}

// Publish an event to the EventSource.
//
// This will update the materialised view and send the event on the "Events" channel. The event will be updated with the
// materialised view.
//
// This is mostly useful in conjunction with NewUnattached, for testing.
func (e EventSource) Publish(event Event) {
	switch event := event.(type) {
	case EventRemove:
		clone := reflect.DeepCopy(e.CanonicalView())
		clone.Modules = slices.DeleteFunc(clone.Modules, func(m *schema.Module) bool {
			if deploymentKey, ok := event.Deployment.Get(); ok {
				return m.Runtime != nil && m.Runtime.Deployment != nil && m.Runtime.Deployment.DeploymentKey == deploymentKey
			}
			return m.Name == event.Module
		})
		e.view.Store(clone)

	case EventUpsert:
		clone := reflect.DeepCopy(e.CanonicalView())
		changeset := reflect.DeepCopy(e.ActiveChangeset())
		var modules []*schema.Module
		if changesetKey, ok := event.Changeset.Get(); ok {
			changeset, ok := changeset.Get()
			if !ok || changeset.Key != changesetKey {
				// Upsert is for a non active changeset
				return
			}
			modules = changeset.Modules
		} else {
			modules = clone.Modules
		}
		if i := slices.IndexFunc(modules, func(m *schema.Module) bool { return m.Name == event.Module.Name }); i != -1 {
			modules[i] = event.Module
		} else {
			modules = append(modules, event.Module)
		}
		if _, ok := event.Changeset.Get(); ok {
			changeset.MustGet().Modules = modules
		} else {
			clone.Modules = modules
		}
		e.view.Store(clone)
		e.activeChangeset.Store(changeset)

	case EventChangesetStarted:
		if event.Changeset.State == schema.ChangesetStatePreparing {
			e.activeChangeset.Store(optional.Some(event.Changeset))
		}
	case EventChangesetEnded:
		activeChangeset, ok := e.activeChangeset.Load().Get()
		if !ok || activeChangeset.Key != event.Key {
			break
		}
		if !event.Success {
			e.activeChangeset.Store(optional.None[*schema.Changeset]())
			break
		}
		allDeps := islices.Map(islices.Filter(e.view.Load().Modules, func(module *schema.Module) bool { return !module.Builtin }), func(m *schema.Module) key.Deployment { return m.Runtime.Deployment.DeploymentKey })
		final := latestSchema(e.CanonicalView(), optional.Some(activeChangeset))
		removedDeps := islices.Filter(allDeps, func(d key.Deployment) bool { return slices.Index(event.ReplacedDeloyments, d) == -1 })
		event.ReplacedDeloyments = removedDeps
		e.view.Store(final)
	}
	event = event.finalize(e.CanonicalView(), e.ActiveChangeset())
	e.events <- event
}

// New creates a new EventSource that pulls schema changes from the SchemaService into an event channel and a
// materialised view (ie. [schema.Schema]).
//
// The sync will terminate when the context is cancelled.
func New(ctx context.Context, client ftlv1connect.SchemaServiceClient) EventSource {
	logger := log.FromContext(ctx).Scope("schema-sync")
	out := NewUnattached()
	more := true
	initialSyncComplete := false
	logger.Debugf("Starting schema pull")

	// Set the initial "live" state by pinging the server. After that we'll rely on the stream.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := client.Ping(pingCtx, connect.NewRequest(&ftlv1.PingRequest{}))
	out.live.Store(err == nil && resp.Msg.NotReady == nil)

	go rpc.RetryStreamingServerStream(ctx, "schema-sync", backoff.Backoff{}, &ftlv1.PullSchemaRequest{}, client.PullSchema, func(_ context.Context, resp *ftlv1.PullSchemaResponse) error {
		out.live.Store(true)
		// resp.More can become true again if the streaming client reconnects, but we don't want downstream to have to
		// care about a new initial sync restarting.
		more = more && resp.More

		switch event := resp.Event.(type) {
		case *ftlv1.PullSchemaResponse_ChangesetCreated_:
			changeset, err := schema.ChangesetFromProto(event.ChangesetCreated.Changeset)
			if err != nil {
				return fmt.Errorf("invalid changeset: %w", err)
			}
			out.Publish(EventChangesetStarted{
				Changeset: changeset,
				more:      more,
			})
		case *ftlv1.PullSchemaResponse_ChangesetFailed_:
			key, err := key.ParseChangesetKey(event.ChangesetFailed.Key)
			if err != nil {
				return fmt.Errorf("invalid changeset key: %w", err)
			}
			out.Publish(EventChangesetEnded{
				Key:     key,
				Success: false,
				Error:   event.ChangesetFailed.Error,
				more:    more,
			})
		case *ftlv1.PullSchemaResponse_ChangesetCommitted_:
			key, err := key.ParseChangesetKey(event.ChangesetCommitted.Changeset.Key)
			if err != nil {
				return fmt.Errorf("invalid changeset key: %w", err)
			}
			out.Publish(EventChangesetEnded{
				Key:     key,
				Success: true,
				more:    more,
			})
		case *ftlv1.PullSchemaResponse_DeploymentCreated_:
			module, err := schema.ValidatedModuleFromProto(event.DeploymentCreated.Schema)
			if err != nil {
				return fmt.Errorf("invalid module: %w", err)
			}
			logger.Tracef("Module %s upserted", module.Name)
			cs := optional.Option[key.Changeset]{}
			if event.DeploymentCreated.Changeset != nil {
				parsed, err := key.ParseChangesetKey(*event.DeploymentCreated.Changeset)
				if err != nil {
					return fmt.Errorf("invalid changeset key: %w", err)
				}
				cs = optional.Some(parsed)
			}

			out.Publish(EventUpsert{
				Changeset: cs,
				Module:    module,
				more:      more,
			})
		case *ftlv1.PullSchemaResponse_DeploymentUpdated_:
			module, err := schema.ValidatedModuleFromProto(event.DeploymentUpdated.Schema)
			if err != nil {
				return fmt.Errorf("invalid module: %w", err)
			}
			var changeset optional.Option[key.Changeset]
			if event.DeploymentUpdated.GetChangeset() != "" {
				c, err := key.ParseChangesetKey(event.DeploymentUpdated.GetChangeset())
				if err != nil {
					return fmt.Errorf("invalid changeset key: %w", err)
				}
				changeset = optional.Some(c)
			}
			logger.Tracef("Module %s upserted", module.Name)
			out.Publish(EventUpsert{
				Module:    module,
				Changeset: changeset,
				more:      more,
			})
		case *ftlv1.PullSchemaResponse_DeploymentRemoved_:
			// TODO: bring this back? but we can't right because there can be multiple?
			// if !resp.ModuleRemoved {
			// 	return nil
			// }

			var deploymentKey optional.Option[key.Deployment]
			if event.DeploymentRemoved.GetKey() != "" {
				k, err := key.ParseDeploymentKey(event.DeploymentRemoved.GetKey())
				if err != nil {
					return fmt.Errorf("invalid deployment key: %w", err)
				}
				deploymentKey = optional.Some(k)
			}
			logger.Debugf("Deployment %s removed", optional.Ptr(event.DeploymentRemoved.Key).Default(event.DeploymentRemoved.ModuleName))
			out.Publish(EventRemove{
				Deployment: deploymentKey,
				Module:     event.DeploymentRemoved.ModuleName,
				more:       more,
			})
		default:
			return fmt.Errorf("schema-sync: unknown change type %T", event)
		}

		if !more && !initialSyncComplete {
			initialSyncComplete = true
			close(out.initialSyncComplete)
		}
		return nil
	}, func(_ error) bool {
		out.live.Store(false)
		return true
	})
	return out
}
