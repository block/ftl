package buildengine

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	enginepbconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/builderrors"
	ftlerrors "github.com/block/ftl/common/errors"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

type updatesService struct {
	engine *Engine
}

var _ enginepbconnect.BuildEngineServiceHandler = &updatesService{}

func (e *Engine) startUpdatesService(ctx context.Context, endpoint *url.URL) error {
	logger := log.FromContext(ctx).Scope("build:updates")

	svc := &updatesService{
		engine: e,
	}

	logger.Debugf("Build updates service listening on: %s", endpoint)
	err := rpc.Serve(ctx, endpoint,
		rpc.GRPC(enginepbconnect.NewBuildEngineServiceHandler, svc),
	)

	if err != nil {
		return fmt.Errorf("build updates service stopped serving: %w", err)
	}
	return nil
}

func (u *updatesService) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (u *updatesService) StreamEngineEvents(ctx context.Context, req *connect.Request[buildenginepb.StreamEngineEventsRequest], stream *connect.ServerStream[buildenginepb.StreamEngineEventsResponse]) error {
	logger := log.FromContext(ctx).Scope("build:updates")
	events := make(chan EngineEvent, 64)
	u.engine.EngineUpdates.Subscribe(events)
	defer u.engine.EngineUpdates.Unsubscribe(events)

	for event := range channels.IterContext(ctx, events) {
		var pbEvent *buildenginepb.EngineEvent
		switch e := event.(type) {
		case EngineStarted:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_EngineStarted{
					EngineStarted: &buildenginepb.EngineStarted{},
				},
			}
		case EngineEnded:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_EngineEnded{
					EngineEnded: &buildenginepb.EngineEnded{
						ModuleErrors: moduleErrorsToErrorList(e.ModuleErrors),
					},
				},
			}
		case ModuleAdded:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleAdded{
					ModuleAdded: &buildenginepb.ModuleAdded{
						Module: e.Module,
					},
				},
			}

		case ModuleRemoved:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleRemoved{
					ModuleRemoved: &buildenginepb.ModuleRemoved{
						Module: e.Module,
					},
				},
			}

		case ModuleBuildWaiting:
			proto, err := langpb.ModuleConfigToProto(e.Config.Abs())
			if err != nil {
				logger.Errorf(err, "failed to marshal module config")
				continue
			}
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
					ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
						Config: proto,
					},
				},
			}

		case ModuleBuildStarted:
			{
				proto, err := langpb.ModuleConfigToProto(e.Config.Abs())
				if err != nil {
					logger.Errorf(err, "failed to marshal module config")
					continue
				}
				pbEvent = &buildenginepb.EngineEvent{
					Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
						ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
							Config:        proto,
							IsAutoRebuild: e.IsAutoRebuild,
						},
					},
				}
			}
		case ModuleBuildSuccess:
			proto, err := langpb.ModuleConfigToProto(e.Config.Abs())
			if err != nil {
				logger.Errorf(err, "failed to marshal module config")
				continue
			}
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
					ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
						Config: proto,
					},
				},
			}

		case ModuleBuildFailed:
			proto, err := langpb.ModuleConfigToProto(e.Config.Abs())
			if err != nil {
				logger.Errorf(err, "failed to marshal module config")
				continue
			}
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
					ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
						Config: proto,
						Errors: &langpb.ErrorList{
							Errors: errorToLangError(e.Error),
						},
						IsAutoRebuild: e.IsAutoRebuild,
					},
				},
			}

		case ModuleDeployStarted:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
					ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
						Module: e.Module,
					},
				},
			}

		case ModuleDeploySuccess:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
					ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
						Module: e.Module,
					},
				},
			}

		case ModuleDeployFailed:
			pbEvent = &buildenginepb.EngineEvent{
				Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
					ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
						Module: e.Module,
						Errors: &langpb.ErrorList{
							Errors: errorToLangError(e.Error),
						},
					},
				},
			}
		}

		if pbEvent != nil {
			err := stream.Send(&buildenginepb.StreamEngineEventsResponse{
				Event: pbEvent,
			})
			if err != nil {
				return fmt.Errorf("failed to send engine event: %w", err)
			}
		}
	}

	return nil
}

// moduleErrorsToErrorList converts a map of module errors to protobuf ErrorList format
func moduleErrorsToErrorList(moduleErrors map[string]error) map[string]*langpb.ErrorList {
	if moduleErrors == nil {
		return nil
	}

	result := make(map[string]*langpb.ErrorList)
	for module, err := range moduleErrors {
		errs := errorToLangError(err)
		if len(errs) > 0 {
			result[module] = &langpb.ErrorList{
				Errors: errs,
			}
		}
	}
	return result
}

// errorToLangError converts a single error to a slice of protobuf Error messages
func errorToLangError(err error) []*langpb.Error {
	errs := []*langpb.Error{}

	// Unwrap and deduplicate all nested errors
	for _, e := range ftlerrors.DeduplicateErrors(ftlerrors.UnwrapAll(err)) {
		if !ftlerrors.Innermost(e) {
			continue
		}

		var berr builderrors.Error
		if !errors.As(e, &berr) {
			// If not a builderrors.Error, create a basic error
			errs = append(errs, &langpb.Error{
				Msg:   e.Error(),
				Level: langpb.Error_ErrorLevel(berr.Level),
			})
			continue
		}

		if berr.Type == builderrors.COMPILER {
			continue
		}

		pbError := &langpb.Error{
			Msg:   berr.Msg,
			Level: langpb.Error_ErrorLevel(berr.Level),
		}

		// Add position information if available
		if pos, ok := berr.Pos.Get(); ok {
			pbError.Pos = &langpb.Position{
				Filename:    pos.Filename,
				Line:        int64(pos.Line),
				StartColumn: int64(pos.StartColumn),
				EndColumn:   int64(pos.EndColumn),
			}
		}

		errs = append(errs, pbError)
	}

	return errs
}
