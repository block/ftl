package dev

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"sync"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/container"
)

//go:embed docker-compose.redpanda.yml
var redpandaDockerCompose string

// use this lock while checking redPandaRunning status and running `docker compose up` if needed
var redPandaLock = &sync.Mutex{}
var redPandaRunning bool

func SetUpRedPanda(ctx context.Context) error {
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
	err := container.ComposeUp(ctx, "redpanda", redpandaDockerCompose, profile)
	if err != nil {
		return fmt.Errorf("could not start redpanda: %w", err)
	}
	redPandaRunning = true
	return nil
}
