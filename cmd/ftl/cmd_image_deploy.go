package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
)

type imageDeployCmd struct {
	Images              []string `arg:"" help:"The images to deploy"`
	AllowInsecureImages bool     `help:"Allows the use of insecure HTTP based registries." env:"FTL_IMAGE_REPOSITORY_ALLOW_INSECURE"`
}

func (b *imageDeployCmd) Run(
	ctx context.Context,
	admin adminpbconnect.AdminServiceClient,
) error {
	return deployImages(ctx, admin, b.Images, b.AllowInsecureImages)
}

func deployImages(ctx context.Context,
	admin adminpbconnect.AdminServiceClient, images []string, allowInsecure bool) error {
	logger := log.FromContext(ctx)
	resp, err := admin.DeployImages(ctx, connect.NewRequest(&adminpb.DeployImagesRequest{Image: images, AllowInsecure: allowInsecure}))
	if err != nil {
		return errors.Wrapf(err, "failed to deploy images %v", images)
	}
	for resp.Receive() {
		// ignore the response, we just want to deploy the images
	}
	if resp.Err() != nil {
		return errors.Wrapf(resp.Err(), "failed to deploy images %v", images)
	}

	logger.Infof("Deployed images %v", images) //nolint
	return nil
}
