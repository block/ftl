package devstate

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	languagepb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

type DevState struct {
	Modules []ModuleState
	Schema  *schema.Schema
}

type ModuleState struct {
	Name   string
	Path   string
	Errors []builderrors.Error
}

// WaitForDevState waits for the engine to finish and prints a summary of the current module (paths and errors) and schema.
// It is useful for synchronizing with FTL dev as it accommodates delays in detecting code changes, building and deploying.
func WaitForDevState(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, schemaClient adminpbconnect.AdminServiceClient, waitForBuildsToStart bool) (DevState, error) {
	stream, err := buildEngineClient.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{
		ReplayHistory: true,
	}))
	if err != nil {
		return DevState{}, errors.Wrap(err, "failed to stream engine events")
	}

	start := time.Now()
	var idleDuration time.Duration
	if waitForBuildsToStart {
		idleDuration = 1_300 * time.Millisecond
	}
	engineEnded := optional.Some(&buildenginepb.EngineEnded{})
	endOfHistoryEvent := optional.None[*buildenginepb.EngineEvent]()

	streamChan := make(chan *buildenginepb.EngineEvent)
	errChan := make(chan error)
	go func() {
		for {
			if !stream.Receive() {
				close(streamChan)
				errChan <- stream.Err()
				return
			}
			streamChan <- stream.Msg().Event
		}
	}()
streamLoop:
	for {
		// We want to wait for code changes in the module to be detected and for builds to start.
		// So we wait for the engine to be idle after idleDuration.
		var idleDeadline <-chan time.Time
		if engineEnded.Ok() && endOfHistoryEvent.Ok() {
			idleDeadline = time.After(max(idleDuration-time.Since(start), time.Millisecond*100-time.Since(endOfHistoryEvent.MustGet().Timestamp.AsTime())))
		}
		select {
		case <-idleDeadline:
			break streamLoop
		case <-ctx.Done():
			return DevState{}, errors.Wrap(ctx.Err(), "did not complete build engine update stream")
		case e, ok := <-streamChan:
			if !ok {
				err = <-errChan
				if errors.Is(err, context.Canceled) {
					return DevState{}, errors.WithStack(ctx.Err()) // nolint:wrapcheck
				}
				return DevState{}, errors.Wrap(err, "failed to stream engine events")
			}

			switch event := e.Event.(type) {
			case *buildenginepb.EngineEvent_ReachedEndOfHistory:
				endOfHistoryEvent = optional.Some(e)
			case *buildenginepb.EngineEvent_EngineStarted:
				engineEnded = optional.None[*buildenginepb.EngineEnded]()
			case *buildenginepb.EngineEvent_EngineEnded:
				engineEnded = optional.Some(event.EngineEnded)

			default:
			}
		}
	}

	engineEndedEvent, ok := engineEnded.Get()
	if !ok {
		return DevState{}, errors.WithStack(errors.New("engine did not end"))
	}

	schemaResp, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return DevState{}, errors.Wrap(err, "failed to get schema")
	}
	sch, err := schema.FromProto(schemaResp.Msg.Schema)
	if err != nil {
		return DevState{}, errors.Wrap(err, "failed to parse schema when waiting for dev state")
	}
	modulesPaths := map[string]string{}
	for _, module := range engineEndedEvent.Modules {
		modulesPaths[module.Module] = module.Path
	}
	out := DevState{}
	out.Modules = slices.Map(engineEndedEvent.Modules, func(m *buildenginepb.EngineEnded_Module) ModuleState {
		return ModuleState{
			Name:   m.Module,
			Path:   m.Path,
			Errors: languagepb.ErrorsFromProto(m.Errors),
		}
	})
	out.Schema = sch
	return out, nil
}
