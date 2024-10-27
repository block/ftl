package relay

import (
	"context"
	"fmt"
	"os"
	"time"

	"ftl/origin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

var logFile = ftl.Config[string]("log_file")
var db = ftl.PostgresDatabase("exemplardb")

// PubSub

var _ = ftl.Subscription(origin.AgentBroadcast, "agentConsumer")

//ftl:subscribe agentConsumer
func ConsumeAgentBroadcast(ctx context.Context, agent origin.Agent, client BriefedClient) error {
	ftl.LoggerFromContext(ctx).Infof("Received agent %v", agent.Id)
	return client(ctx, agent)
}

type AgentDeployment struct {
	Agent  origin.Agent
	Target string
}

type MissionSuccess struct {
	AgentID   int
	SuccessAt time.Time
}

type AgentTerminated struct {
	AgentID      int
	TerminatedAt time.Time
}

//ftl:verb
func Briefed(ctx context.Context, agent origin.Agent, deployed DeployedClient) error {
	briefedAt := time.Now()
	ftl.LoggerFromContext(ctx).Infof("Briefed agent %v at %s", agent.Id, briefedAt)
	agent.BriefedAt = ftl.Some(briefedAt)
	d := AgentDeployment{
		Agent:  agent,
		Target: "villain",
	}
	return deployed(ctx, d)
}

//ftl:verb
func Deployed(ctx context.Context, d AgentDeployment) error {
	ftl.LoggerFromContext(ctx).Infof("Deployed agent %v to %s", d.Agent.Id, d.Target)
	return appendLog(ctx, "deployed %d", d.Agent.Id)
}

//ftl:verb
func Succeeded(ctx context.Context, s MissionSuccess) error {
	ftl.LoggerFromContext(ctx).Infof("Agent %d succeeded at %s\n", s.AgentID, s.SuccessAt)
	return appendLog(ctx, "succeeded %d", s.AgentID)
}

//ftl:verb
func Terminated(ctx context.Context, t AgentTerminated) error {
	ftl.LoggerFromContext(ctx).Infof("Agent %d terminated at %s\n", t.AgentID, t.TerminatedAt)
	return appendLog(ctx, "terminated %d", t.AgentID)
}

// Exported verbs

type MissionResultRequest struct {
	AgentID    int
	Successful bool
}

type MissionResultResponse struct{}

//ftl:verb export
func MissionResult(ctx context.Context, req MissionResultRequest, success SucceededClient, failure TerminatedClient) (MissionResultResponse, error) {
	ftl.LoggerFromContext(ctx).Infof("Mission result for agent %v: %t\n", req.AgentID, req.Successful)
	agentID := req.AgentID
	var event any
	if req.Successful {
		event = MissionSuccess{
			AgentID:   int(agentID),
			SuccessAt: time.Now(),
		}
		err := success(ctx, event.(MissionSuccess))
		if err != nil {
			return MissionResultResponse{}, err
		}
	} else {
		event = AgentTerminated{
			AgentID:      int(agentID),
			TerminatedAt: time.Now(),
		}
		err := failure(ctx, event.(AgentTerminated))
		if err != nil {
			return MissionResultResponse{}, err
		}
	}
	ftl.LoggerFromContext(ctx).Infof("Sending event %v\n", event)
	return MissionResultResponse{}, nil
}

type GetLogFileRequest struct{}
type GetLogFileResponse struct {
	Path string
}

//ftl:verb export
func GetLogFile(ctx context.Context, req GetLogFileRequest) (GetLogFileResponse, error) {
	return GetLogFileResponse{Path: logFile.Get(ctx)}, nil
}

// Helpers

func appendLog(ctx context.Context, msg string, args ...interface{}) error {
	path := logFile.Get(ctx)
	if path == "" {
		return fmt.Errorf("log_file config not set")
	}
	w, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %q: %w", path, err)
	}
	fmt.Fprintf(w, msg+"\n", args...)
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file %q: %w", path, err)
	}
	return nil
}
