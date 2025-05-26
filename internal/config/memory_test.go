package config_test

import (
	"testing"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/config"
)

func TestMemory(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(t.Context())
	testConfig(t, ctx, config.NewMemoryProvider[config.Configuration]())
}
