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

	imageService, err := oci.NewImageService(ctx, &b.ImageConfig)
	if err != nil {
		return errors.Wrapf(err, "failed to init OCI")
	}

	sch, _, err := imageService.PullSchema(ctx, b.Image)
	if err != nil {
		return errors.Wrapf(err, "failed to pull image schema %s", b.Image)
	}
	logger.Infof("%s", sch.String()) //nolint
	return nil
}
