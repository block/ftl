package timelineclient

import (
	"context"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/channels"
)

// LogSink is a log sink that sends logs to the timeline client.
//
// It needs to be run in a separate goroutine after creation by calling RunLogLoop.
type LogSink struct {
	client   *Client
	logQueue chan log.Entry
	level    log.Level
}

var _ log.Sink = (*LogSink)(nil)

func NewLogSink(client *Client, level log.Level) *LogSink {
	return &LogSink{
		client:   client,
		logQueue: make(chan log.Entry, 10000),
		level:    level,
	}
}

// Log implements Sink
func (l *LogSink) Log(entry log.Entry) error {
	if entry.Level < l.level {
		return nil
	}

	select {
	case l.logQueue <- entry:
	default:
		// Drop log entry if queue is full
		return errors.Errorf("log queue is full")
	}
	return nil
}

// RunLogLoop runs the log loop.
//
// It will run until the context is cancelled.
func (l *LogSink) RunLogLoop(ctx context.Context) {
	for entry := range channels.IterContext(ctx, l.logQueue) {
		dep, ok := entry.Attributes["deployment"]
		var deploymentKey key.Deployment
		var err error
		if ok {
			deploymentKey, err = key.ParseDeploymentKey(dep)
			if err != nil {
				continue
			}
		}

		cs, ok := entry.Attributes["changeset"]
		var changesetKey optional.Option[key.Changeset]
		if ok {
			k, err := key.ParseChangesetKey(cs)
			if err != nil {
				continue
			}
			changesetKey = optional.Some(k)
		}

		var errorString *string
		if entry.Error != nil {
			errStr := entry.Error.Error()
			errorString = &errStr
		}
		var request optional.Option[key.Request]
		if reqStr, ok := entry.Attributes["request"]; ok {
			req, err := key.ParseRequestKey(reqStr) //nolint:errcheck // best effort
			if err == nil {
				request = optional.Some(req)
			}
		}
		l.client.Publish(ctx, &Log{
			DeploymentKey: deploymentKey,
			RequestKey:    request,
			Level:         int32(entry.Level), //nolint:gosec
			Time:          entry.Time,
			Attributes:    entry.Attributes,
			Message:       entry.Message,
			Error:         optional.Ptr(errorString),
			ChangesetKey:  changesetKey,
		})
	}
}
