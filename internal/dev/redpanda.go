package dev

import (
	"context"
	_ "embed"
	"os"
	"sync"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/container"
	"github.com/block/ftl/internal/projectconfig"
)

//go:embed docker-compose.redpanda.yml
var redpandaDockerCompose string

// use this lock while checking redPandaRunning status and running `docker compose up` if needed
var redPandaLock = &sync.Mutex{}
var redPandaRunning bool

func SetUpRedPanda(ctx context.Context, project projectconfig.Config) error {
	redPandaLock.Lock()
	defer redPandaLock.Unlock()

	if redPandaRunning {
		return nil
	}
	var profile optional.Option[string]
	if _, ci := os.LookupEnv("CI"); !ci {
		// include console except in CI
		profile = optional.Some[string]("console")
	}
	_, err := container.ComposeUp(ctx, project, "redpanda", redpandaDockerCompose, profile)
	if err != nil {
		return errors.Wrap(err, "could not start redpanda")
	}
	redPandaRunning = true
	return nil
}
