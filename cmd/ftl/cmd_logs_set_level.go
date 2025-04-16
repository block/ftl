package main

import (
	"context"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/internal/log"
)

type logsSetLevelCmd struct {
	Level string `arg:"" help:"The new log level to set." predictor:"log-level"`
}

func (k *logsSetLevelCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	level, err := log.LevelString(k.Level)
	if err != nil {
		return errors.Wrap(err, "invalid log level")
	}
	log.SetCurrentLevel(logger, level)
	log.ReplayLogs(ctx)
	return nil
}
