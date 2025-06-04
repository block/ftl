package main

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/jpillora/backoff"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/internal/rpc"
)

type pingCmd struct {
	Wait time.Duration `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1s"`
}

func (c *pingCmd) Run(ctx context.Context, controller adminpbconnect.AdminServiceClient) error {
	return errors.WithStack(rpc.Wait(ctx, backoff.Backoff{Max: time.Second}, c.Wait, controller)) //nolint:wrapcheck
}
