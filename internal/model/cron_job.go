package model

import (
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

type CronJob struct {
	Key           key.CronJob
	DeploymentKey key.Deployment
	Verb          schema.Ref
	Schedule      string
	StartTime     time.Time
	NextExecution time.Time
	LastExecution optional.Option[time.Time]
}
