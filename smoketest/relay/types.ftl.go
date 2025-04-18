// Code generated by FTL. DO NOT EDIT.
package relay

import (
	"context"
	ftlorigin "ftl/origin"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/server"
)

type AppendLogClient func(context.Context, AppendLogRequest) error

type BriefedClient func(context.Context, ftlorigin.Agent) error

type DeployedClient func(context.Context, AgentDeployment) error

type ConsumeAgentBroadcastClient func(context.Context, ftlorigin.Agent) error

type FetchLogsClient func(context.Context, FetchLogsRequest) (FetchLogsResponse, error)

type MissionResultClient func(context.Context, MissionResultRequest) (MissionResultResponse, error)

type SucceededClient func(context.Context, MissionSuccess) error

type TerminatedClient func(context.Context, AgentTerminated) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			AppendLog,
			server.Config[string]("relay", "logFile"),
		),
		reflection.ProvideResourcesForVerb(
			Briefed,
			server.SinkClient[DeployedClient, AgentDeployment](),
		),
		reflection.ProvideResourcesForVerb(
			Deployed,
			server.Config[string]("relay", "logFile"),
		),
		reflection.ProvideResourcesForVerb(
			ConsumeAgentBroadcast,
			server.SinkClient[BriefedClient, ftlorigin.Agent](),
		),
		reflection.ProvideResourcesForVerb(
			FetchLogs,
			server.Config[string]("relay", "logFile"),
		),
		reflection.ProvideResourcesForVerb(
			MissionResult,
			server.SinkClient[SucceededClient, MissionSuccess](),
			server.SinkClient[TerminatedClient, AgentTerminated](),
		),
		reflection.ProvideResourcesForVerb(
			Succeeded,
			server.Config[string]("relay", "logFile"),
		),
		reflection.ProvideResourcesForVerb(
			Terminated,
			server.Config[string]("relay", "logFile"),
		),
	)
}
