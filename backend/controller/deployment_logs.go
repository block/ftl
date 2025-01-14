package controller

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/timelineclient"
)

var _ log.Sink = (*deploymentLogsSink)(nil)

func newDeploymentLogsSink(ctx context.Context, timelineClient *timelineclient.Client) *deploymentLogsSink {
	sink := &deploymentLogsSink{
		logQueue: make(chan log.Entry, 10000),
	}

	// Process logs in background
	go sink.processLogs(ctx, timelineClient)

	return sink
}

type deploymentLogsSink struct {
	logQueue chan log.Entry
}

// Log implements Sink
func (d *deploymentLogsSink) Log(entry log.Entry) error {
	select {
	case d.logQueue <- entry:
	default:
		// Drop log entry if queue is full
		return fmt.Errorf("log queue is full")
	}
	return nil
}

func (d *deploymentLogsSink) processLogs(ctx context.Context, timelineClient *timelineclient.Client) {
	for entry := range channels.IterContext(ctx, d.logQueue) {
		var deployment key.Deployment
		depStr, ok := entry.Attributes["deployment"]
		if !ok {
			continue
		}

		dep, err := key.ParseDeploymentKey(depStr)
		if err != nil {
			continue
		}
		deployment = dep

		var request optional.Option[key.Request]
		if reqStr, ok := entry.Attributes["request"]; ok {
			req, err := key.ParseRequestKey(reqStr)
			if err == nil {
				request = optional.Some(req)
			}
		}

		var errorStr optional.Option[string]
		if entry.Error != nil {
			errorStr = optional.Some(entry.Error.Error())
		}

		timelineClient.Publish(ctx, &timelineclient.Log{
			RequestKey:    request,
			DeploymentKey: deployment,
			Time:          entry.Time,
			Level:         int32(entry.Level.Severity()),
			Attributes:    entry.Attributes,
			Message:       entry.Message,
			Error:         errorStr,
		})
	}
}
