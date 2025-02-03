package cron

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl"
)

//ftl:cron 2s
func Job(ctx context.Context) error {
	ftl.LoggerFromContext(ctx).Infof("Cron job executed")
	return nil
}
