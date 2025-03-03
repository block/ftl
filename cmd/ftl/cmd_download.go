package main

import (
	"context"
	"fmt"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/download"
	"github.com/block/ftl/internal/key"
)

type downloadCmd struct {
	Dest       string         `short:"d" help:"Destination directory." default:"."`
	Deployment key.Deployment `help:"Deployment to download." arg:"" predictor:"deployments"`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.AdminServiceClient) error {
	err := download.Artefacts(ctx, client, d.Deployment, d.Dest)
	if err != nil {
		return fmt.Errorf("failed to download artefacts: %w", err)
	}
	return nil
}
