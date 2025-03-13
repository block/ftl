package config_test

import (
	"testing"

	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/log"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(t.Context())
	dir := t.TempDir()

	provider := config.NewFileProvider[config.Configuration]("test", dir)
	testConfig(t, ctx, provider)
}
