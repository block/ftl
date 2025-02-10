package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	languagepb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

// awaitSummaryCmd waits for the engine to finish and prints a summary of the current module (paths and errors) and schema.
// It is useful for goose as it accomodates delays in detecting code changes, building and deploying.
type awaitSummaryCmd struct {
}

func (c *awaitSummaryCmd) Run(ctx context.Context, buildEngineClient buildenginepbconnect.BuildEngineServiceClient, schemaClient ftlv1connect.SchemaServiceClient) error {
	stream, err := buildEngineClient.StreamEngineEvents(ctx, connect.NewRequest(&buildenginepb.StreamEngineEventsRequest{
		ReplayHistory: true,
	}))
	if err != nil {
		return fmt.Errorf("failed to stream engine events: %w", err)
	}

	start := time.Now()
	idleDuration := 1500 * time.Millisecond
	engineEnded := optional.None[*buildenginepb.EngineEnded]()

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
			return ctx.Err()
		case event, ok := <-streamChan:
			if !ok {
				err = <-errChan
				if errors.Is(err, context.Canceled) {
					return nil
				}
				return fmt.Errorf("failed to stream engine events: %w", err)
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
		return errors.New("engine did not end")
	}

	schemaResp, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}
	schema, err := schema.FromProto(schemaResp.Msg.Schema)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	fmt.Printf("Module Overview:\n")
	if len(engineEndedEvent.Modules) == 0 {
		fmt.Println("No modules found.")
		return nil
	}
	for _, module := range engineEndedEvent.Modules {
		fmt.Printf("%s (%s)\n", module.Module, module.Path)
		if module.Errors == nil || len(module.Errors.Errors) == 0 {
			fmt.Println("  Success with no warnings.")
			continue
		}
		fmt.Print(strings.Join(slices.Map(module.Errors.Errors, func(e *languagepb.Error) string {
			var errorType string
			switch e.Level {
			case languagepb.Error_ERROR_LEVEL_ERROR:
				errorType = "Error"
			case languagepb.Error_ERROR_LEVEL_WARN:
				errorType = "Warn"
			case languagepb.Error_ERROR_LEVEL_INFO:
				errorType = "Info"
			default:
				panic(fmt.Sprintf("unknown error type: %v", e.Level))
			}
			var posStr string
			if e.Pos != nil {
				posStr = fmt.Sprintf("%s:%d", e.Pos.Filename, e.Pos.StartColumn)
				if e.Pos.EndColumn != e.Pos.StartColumn {
					posStr += fmt.Sprintf(":%d", e.Pos.EndColumn)
				}
				posStr += ": "
			}
			return fmt.Sprintf("  [%s] %s%s", errorType, posStr, e.Msg)
		}), "\n"))
	}

	fmt.Printf("\n\nSchema:\n%s", schema.String())
	return nil
}
