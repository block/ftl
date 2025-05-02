package dev

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/container"
	"github.com/block/ftl/internal/projectconfig"
)

//go:embed docker-compose.registry.yml
var registryDockerCompose string

func SetupRegistry(ctx context.Context, project projectconfig.Config, image string, port int) error {
	_, err := container.ComposeUp(ctx, project, "registry", registryDockerCompose, optional.None[string](),
		"FTL_REGISTRY_IMAGE="+image,
		"FTL_REGISTRY_PORT="+strconv.Itoa(port))
	if err != nil {
		return errors.Wrap(err, "could not start registry")
	}
	err = WaitForPortReady(ctx, port)
	if err != nil {
		return errors.Wrap(err, "registry container failed to be healthy")
	}
	return nil
}

func WaitForPortReady(ctx context.Context, port int) error {
	timeout := time.After(10 * time.Minute)
	retry := time.NewTicker(5 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("context cancelled waiting for container")
		case <-timeout:
			return errors.Errorf("timed out waiting for container to be healthy")
		case <-retry.C:
			url := fmt.Sprintf("http://127.0.0.1:%d", port)

			req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil) //nolint:gosec
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil

			}
		}

	}
}
