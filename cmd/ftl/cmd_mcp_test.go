package main

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/kong"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/profiles"
)

func TestMCPServerCreation(t *testing.T) {
	t.Parallel()

	ctx := log.ContextWithNewDefaultLogger(t.Context())
	csm := &currentStatusManager{}
	k := createKongApplication(&cli, csm)
	assert.NotPanics(t, func() {
		_ = newMCPServer(ctx, k, profiles.ProjectConfig{}, nil, nil, nil, func(ctx context.Context, kctx *kong.Context) context.Context {
			return ctx
		})
	})
}
