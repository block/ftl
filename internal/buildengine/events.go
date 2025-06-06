package buildengine

import (
	errors "github.com/alecthomas/errors"
	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	langpb "github.com/block/ftl/backend/protos/xyz/block/ftl/language/v1"
	"github.com/block/ftl/internal/moduleconfig"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newModuleRemovedEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleRemoved{
			ModuleRemoved: &buildenginepb.ModuleRemoved{
				Module: module,
			},
		},
	}
}

func newModuleBuildWaitingEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildWaiting{
			ModuleBuildWaiting: &buildenginepb.ModuleBuildWaiting{
				Config: proto,
			},
		},
	}, nil
}

func newModuleBuildStartedEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{Event: &buildenginepb.EngineEvent_ModuleBuildStarted{
		ModuleBuildStarted: &buildenginepb.ModuleBuildStarted{
			Config: proto,
		},
	},
	}, nil
}

func newModuleBuildFailedEvent(config moduleconfig.ModuleConfig, buildErr error) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildFailed{
			ModuleBuildFailed: &buildenginepb.ModuleBuildFailed{
				Config: proto,
				Errors: &langpb.ErrorList{Errors: errorToLangError(buildErr)},
			},
		},
	}, nil
}

func newModuleBuildSuccessEvent(config moduleconfig.ModuleConfig) (*buildenginepb.EngineEvent, error) {
	proto, err := langpb.ModuleConfigToProto(config.Abs())
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module config to proto")
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleBuildSuccess{
			ModuleBuildSuccess: &buildenginepb.ModuleBuildSuccess{
				Config: proto,
			},
		},
	}, nil
}

func newModuleDeployWaitingEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployWaiting{
			ModuleDeployWaiting: &buildenginepb.ModuleDeployWaiting{
				Module: module,
			},
		},
	}
}

func newModuleDeployStartedEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployStarted{
			ModuleDeployStarted: &buildenginepb.ModuleDeployStarted{
				Module: module,
			},
		},
	}
}

func newModuleDeployFailedEvent(module string, deployErr error) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeployFailed{
			ModuleDeployFailed: &buildenginepb.ModuleDeployFailed{
				Module: module,
				Errors: &langpb.ErrorList{Errors: errorToLangError(deployErr)},
			},
		},
	}
}

func newModuleDeploySuccessEvent(module string) *buildenginepb.EngineEvent {
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_ModuleDeploySuccess{
			ModuleDeploySuccess: &buildenginepb.ModuleDeploySuccess{
				Module: module,
			},
		},
	}
}

func newEngineEndedEvent(moduleStates map[string]*moduleState) *buildenginepb.EngineEvent {
	modulesOutput := []*buildenginepb.EngineEnded_Module{}
	for name, state := range moduleStates {
		modulesOutput = append(modulesOutput, &buildenginepb.EngineEnded_Module{
			Module: name,
			Path:   state.meta.module.Config.Dir,
			Errors: state.lastEvent.GetModuleBuildFailed().GetErrors(),
		})
	}
	return &buildenginepb.EngineEvent{
		Timestamp: timestamppb.Now(),
		Event: &buildenginepb.EngineEvent_EngineEnded{
			EngineEnded: &buildenginepb.EngineEnded{
				Modules: modulesOutput,
			},
		},
	}
}
