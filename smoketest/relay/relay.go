package relay

import (
	"context"
	"fmt"
	"ftl/origin"
	"os"
	"time"

	errors "github.com/alecthomas/errors" // Import the FTL SDK.

	"github.com/block/ftl/go-runtime/ftl"
)

type LogFile = ftl.Config[string]

//ftl:verb
//ftl:subscribe origin.agentBroadcast from=beginning
func ConsumeAgentBroadcast(ctx context.Context, agent origin.Agent, client BriefedClient) error {
	ftl.LoggerFromContext(ctx).Infof("Received agent %v", agent.Id)
	return errors.WithStack(client(ctx, agent))
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
	ftl.LoggerFromContext(ctx).Infof("Briefed agent %v", agent.Id)
	d := AgentDeployment{
		Agent:  agent,
		Target: "villain",
	}
	return errors.WithStack(deployed(ctx, d))
}

//ftl:verb
func Deployed(ctx context.Context, d AgentDeployment, logFile LogFile) error {
	ftl.LoggerFromContext(ctx).Infof("Deployed agent %v to %s", d.Agent.Id, d.Target)
	return errors.WithStack(appendLog(ctx, logFile, "deployed %d", d.Agent.Id))
}

//ftl:verb
func Succeeded(ctx context.Context, s MissionSuccess, logFile LogFile) error {
	ftl.LoggerFromContext(ctx).Infof("Agent %d succeeded at %s\n", s.AgentID, s.SuccessAt)
	return errors.WithStack(appendLog(ctx, logFile, "succeeded %d", s.AgentID))
}

//ftl:verb
func Terminated(ctx context.Context, t AgentTerminated, logFile LogFile) error {
	ftl.LoggerFromContext(ctx).Infof("Agent %d terminated at %s\n", t.AgentID, t.TerminatedAt)
	return errors.WithStack(appendLog(ctx, logFile, "terminated %d", t.AgentID))
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
		err := success(ctx, event.(MissionSuccess)) //nolint:forcetypeassert
		if err != nil {
			return MissionResultResponse{}, errors.WithStack(err)
		}
	} else {
		event = AgentTerminated{
			AgentID:      int(agentID),
			TerminatedAt: time.Now(),
		}
		err := failure(ctx, event.(AgentTerminated)) //nolint:forcetypeassert
		if err != nil {
			return MissionResultResponse{}, errors.WithStack(err)
		}
	}
	ftl.LoggerFromContext(ctx).Infof("Sending event %v\n", event)
	return MissionResultResponse{}, nil
}

// Logging

type AppendLogRequest struct {
	Message string `json:"message"`
}

type FetchLogsRequest struct{}

type FetchLogsResponse struct {
	Messages []string `json:"messages"`
}

//ftl:verb export
func AppendLog(ctx context.Context, req AppendLogRequest, logFile LogFile) error {
	ftl.LoggerFromContext(ctx).Infof("Appending message: %s", req.Message)
	return errors.WithStack(appendLog(ctx, logFile, req.Message))
}

//ftl:verb export
func FetchLogs(ctx context.Context, req FetchLogsRequest, logFile LogFile) (FetchLogsResponse, error) {
	path := logFile.Get(ctx)
	if path == "" {
		return FetchLogsResponse{}, errors.Errorf("logFile config not set")
	}
	r, err := os.Open(path)
	if err != nil {
		return FetchLogsResponse{}, errors.Wrapf(err, "failed to open log file %q", path)
	}
	defer r.Close()
	var messages []string
	for {
		var msg string
		_, err := fmt.Fscanln(r, &msg)
		if err != nil {
			break
		}
		messages = append(messages, msg)
	}
	return FetchLogsResponse{Messages: messages}, nil
}

// Helpers

func appendLog(ctx context.Context, logFile LogFile, msg string, args ...interface{}) error {
	path := logFile.Get(ctx)
	if path == "" {
		return errors.Errorf("logFile config not set")
	}
	w, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open log file %q", path)
	}
	fmt.Fprintf(w, msg+"\n", args...)
	err = w.Close()
	if err != nil {
		return errors.Wrapf(err, "failed to close log file %q", path)
	}
	return nil
}
