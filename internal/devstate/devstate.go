package devstate

import (
	"context"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
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
func WaitForDevState(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, schemaClient adminpbconnect.AdminServiceClient) (DevState, error) {
	stream, err := buildEngineClient.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{
		ReplayHistory: true,
	}))
	if err != nil {
		return DevState{}, errors.Wrap(err, "failed to stream engine events")
	}

	start := time.Now()
	idleDuration := 1500 * time.Millisecond
	engineEnded := optional.Some(&buildenginepb.EngineEnded{})

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
		if engineEnded.Ok() {
			idleDeadline = time.After(idleDuration - time.Since(start))
		}
		select {
		case <-idleDeadline:
			break streamLoop
		case <-ctx.Done():
			return DevState{}, errors.Wrap(ctx.Err(), "did not complete build engine update stream")
		case event, ok := <-streamChan:
			if !ok {
				err = <-errChan
				if errors.Is(err, context.Canceled) {
					return DevState{}, errors.WithStack(ctx.Err()) // nolint:wrapcheck
				}
				return DevState{}, errors.Wrap(err, "failed to stream engine events")
			}

			switch event := event.Event.(type) {
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
