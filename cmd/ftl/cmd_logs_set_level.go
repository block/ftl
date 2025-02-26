package main

import (
	"context"
	"fmt"

	"github.com/block/ftl/internal/log"
)

type logsSetLevelCmd struct {
	Level string `arg:"" help:"The new log level to set." predictor:"log-level"`
}

func (k *logsSetLevelCmd) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	level, err := log.LevelString(k.Level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	log.SetCurrentLevel(logger, level)
	return nil
}
