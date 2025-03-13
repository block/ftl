package config_test

import (
	"testing"

	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/log"
)

func TestMemory(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(t.Context())
	testConfig(t, ctx, config.NewMemoryProvider[config.Configuration]())
}
