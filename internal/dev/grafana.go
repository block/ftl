package dev

import (
	"context"
	_ "embed"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/container"
	"github.com/block/ftl/internal/projectconfig"
)

//go:embed docker-compose.grafana.yml
var grafanaDockerCompose string

func SetupGrafana(ctx context.Context, project projectconfig.Config, image string) error {
	_, err := container.ComposeUp(ctx, project, "grafana", grafanaDockerCompose, optional.None[string]())
	if err != nil {
		return errors.Wrap(err, "could not start grafana")
	}
	err = WaitForPortReady(ctx, 3000)
	if err != nil {
		return errors.Wrap(err, "registry container failed to be healthy")
	}
	return nil
}
