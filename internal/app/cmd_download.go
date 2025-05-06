package app

import (
	"context"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/key"
)

type downloadCmd struct {
	Dest       string         `short:"d" help:"Destination directory." default:"."`
	Deployment key.Deployment `help:"Deployment to download." arg:"" predictor:"deployments"`
}

func (d *downloadCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	err := download.Artefacts(ctx, client, d.Deployment, d.Dest)
	if err != nil {
		return errors.Wrap(err, "failed to download artefacts")
	}
	return nil
}
