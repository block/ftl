package config_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	ctx := log.ContextWithNewDefaultLogger(t.Context())
	dir := t.TempDir()

	provider, err := config.NewFileProvider[config.Configuration](dir, "config.json")
	assert.NoError(t, err)
	testConfig(t, ctx, provider)
}
