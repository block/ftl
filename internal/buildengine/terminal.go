package buildengine

import (
	"context"

	"github.com/alecthomas/types/pubsub"

	buildenginepb "github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/terminal"
)

func updateTerminalWithEngineEvents(ctx context.Context, topic *pubsub.Topic[*buildenginepb.EngineEvent]) {
	events := make(chan *buildenginepb.EngineEvent, 64)
	topic.Subscribe(events)

	go func() {
		defer topic.Unsubscribe(events)
		for event := range channels.IterContext(ctx, events) {
			switch evt := event.Event.(type) {
			case *buildenginepb.EngineEvent_EngineStarted:
			case *buildenginepb.EngineEvent_EngineEnded:

			case *buildenginepb.EngineEvent_ModuleAdded:
				terminal.UpdateModuleState(ctx, evt.ModuleAdded.Module, terminal.BuildStateWaiting)
			case *buildenginepb.EngineEvent_ModuleRemoved:
				terminal.UpdateModuleState(ctx, evt.ModuleRemoved.Module, terminal.BuildStateTerminated)

			case *buildenginepb.EngineEvent_ModuleBuildWaiting:
				terminal.UpdateModuleState(ctx, evt.ModuleBuildWaiting.Config.Name, terminal.BuildStateWaiting)
			case *buildenginepb.EngineEvent_ModuleBuildStarted:
				terminal.UpdateModuleState(ctx, evt.ModuleBuildStarted.Config.Name, terminal.BuildStateBuilding)
			case *buildenginepb.EngineEvent_ModuleBuildSuccess:
				terminal.UpdateModuleState(ctx, evt.ModuleBuildSuccess.Config.Name, terminal.BuildStateBuilt)
			case *buildenginepb.EngineEvent_ModuleBuildFailed:
				terminal.UpdateModuleState(ctx, evt.ModuleBuildFailed.Config.Name, terminal.BuildStateFailed)

			case *buildenginepb.EngineEvent_ModuleDeployStarted:
				terminal.UpdateModuleState(ctx, evt.ModuleDeployStarted.Module, terminal.BuildStateDeploying)
			case *buildenginepb.EngineEvent_ModuleDeploySuccess:
				terminal.UpdateModuleState(ctx, evt.ModuleDeploySuccess.Module, terminal.BuildStateDeployed)
			case *buildenginepb.EngineEvent_ModuleDeployFailed:
				terminal.UpdateModuleState(ctx, evt.ModuleDeployFailed.Module, terminal.BuildStateFailed)
			}
		}
	}()
}
