package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"

	ftlencryption "github.com/block/ftl/backend/controller/encryption/api"
	"github.com/block/ftl/backend/controller/timeline/internal/sql"
	"github.com/block/ftl/backend/libdal"
	"github.com/block/ftl/internal/model"
	"github.com/block/ftl/internal/schema"
)

type CronScheduledEvent struct {
	ID       int64
	Duration time.Duration
	CronScheduled
}

func (e *CronScheduledEvent) GetID() int64 { return e.ID }
func (e *CronScheduledEvent) event()       {}

type CronScheduled struct {
	DeploymentKey model.DeploymentKey
	Verb          schema.Ref

	Time        time.Time
	ScheduledAt time.Time
	Schedule    string
	Error       optional.Option[string]
}

func (e *CronScheduled) toEvent() (Event, error) { //nolint:unparam
	return &CronScheduledEvent{
		CronScheduled: *e,
		Duration:      time.Since(e.Time),
	}, nil
}

type eventCronScheduledJSON struct {
	DurationMS  int64                   `json:"duration_ms"`
	ScheduledAt time.Time               `json:"scheduled_at"`
	Schedule    string                  `json:"schedule"`
	Error       optional.Option[string] `json:"error,omitempty"`
}

func (s *Service) insertCronScheduledEvent(ctx context.Context, querier sql.Querier, event *CronScheduledEvent) error {
	cronJSON := eventCronScheduledJSON{
		DurationMS:  event.Duration.Milliseconds(),
		ScheduledAt: event.ScheduledAt,
		Schedule:    event.Schedule,
		Error:       event.Error,
	}

	data, err := json.Marshal(cronJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal cron JSON: %w", err)
	}

	var payload ftlencryption.EncryptedTimelineColumn
	err = s.encryption.EncryptJSON(json.RawMessage(data), &payload)
	if err != nil {
		return fmt.Errorf("failed to encrypt cron JSON: %w", err)
	}

	err = libdal.TranslatePGError(querier.InsertTimelineCronScheduledEvent(ctx, sql.InsertTimelineCronScheduledEventParams{
		DeploymentKey: event.DeploymentKey,
		TimeStamp:     event.Time,
		Module:        event.Verb.Module,
		Verb:          event.Verb.Name,
		Payload:       payload,
	}))
	if err != nil {
		return fmt.Errorf("failed to insert cron event: %w", err)
	}
	return err
}
