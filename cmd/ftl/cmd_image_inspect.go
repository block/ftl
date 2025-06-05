package main

import (
	"context"

	"github.com/alecthomas/errors"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/oci"
)

type imageInspectCmd struct {
	ImageConfig oci.ImageConfig `embed:""`
	Image       string          `arg:"" help:"The image to inspect"`
}

func (b *imageInspectCmd) Run(
	ctx context.Context,
) error {
	logger := log.FromContext(ctx)

	imageService, err := oci.NewImageService(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to init OCI")
	}

	ref, err := imageService.ParseName(b.Image, b.ImageConfig.AllowInsecureImages)
	if err != nil {
		return errors.Wrapf(err, "failed to parse image name %s", b.Image)
	}

	sch, _, err := imageService.PullSchema(ctx, ref)
	if err != nil {
		return errors.Wrapf(err, "failed to pull image schema %s", b.Image)
	}
	logger.Infof("%s", sch.String()) //nolint
	return nil
}
