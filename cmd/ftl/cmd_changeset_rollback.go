package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

type rollbackChangesetCmd struct {
	Changeset string `arg:"" help:"Changeset to rollback."`
	Force     bool   `help:"Force rollback without de-provisioning, can only be used if the changeset is all ready rolling back and has stalled."`
}

func (g *rollbackChangesetCmd) Run(ctx context.Context, client ftlv1connect.AdminServiceClient) error {
	if g.Force {
		_, err := client.FailChangeset(ctx, connect.NewRequest(&ftlv1.FailChangesetRequest{Changeset: g.Changeset}))
		if err != nil {
			return fmt.Errorf("failed to force rollback changeset: %w", err)
		}
	} else {
		_, err := client.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: g.Changeset}))
		if err != nil {
			return fmt.Errorf("failed to rollback changeset: %w", err)
		}
	}
	return nil
}
