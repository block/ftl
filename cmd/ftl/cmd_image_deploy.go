package main

import (
	"connectrpc.com/connect"
	"context"
	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
)

type imageDeployCmd struct {
	Images []string `arg:"" help:"The images to deploy"`
}

func (b *imageDeployCmd) Run(
	ctx context.Context,
	admin adminpbconnect.AdminServiceClient,
) error {
	logger := log.FromContext(ctx)
	resp, err := admin.DeployImages(ctx, connect.NewRequest(&adminpb.DeployImagesRequest{Image: b.Images}))
	if err != nil {
		return errors.Wrapf(err, "failed to deploy images %v", b.Images)
	}
	for resp.Receive() {
		// ignore the response, we just want to deploy the images
	}
	if resp.Err() != nil {
		return errors.Wrapf(resp.Err(), "failed to deploy images %v", b.Images)
	}

	logger.Infof("Deployed images %v", b.Images) //nolint
	return nil
}
