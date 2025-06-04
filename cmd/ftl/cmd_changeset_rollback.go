package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
)

type rollbackChangesetCmd struct {
	Changeset string `arg:"" help:"Changeset to rollback."`
	Force     bool   `help:"Force rollback without de-provisioning, can only be used if the changeset is all ready rolling back and has stalled."`
}

func (g *rollbackChangesetCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	if g.Force {
		_, err := client.FailChangeset(ctx, connect.NewRequest(&ftlv1.FailChangesetRequest{Changeset: g.Changeset}))
		if err != nil {
			return errors.Wrap(err, "failed to force rollback changeset")
		}
	} else {
		_, err := client.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: g.Changeset}))
		if err != nil {
			return errors.Wrap(err, "failed to rollback changeset")
		}
	}
	return nil
}
