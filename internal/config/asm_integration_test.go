//go:build integration

package config_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/container"
	"github.com/block/ftl/internal/log"
)

//go:embed testdata/docker-compose.yml
var composeYAML string

func TestASMIntegration(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.TODO())
	down, err := container.ComposeUp(ctx, "asm", composeYAML, optional.None[string]())
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, down()) })

	client := secretsmanager.New(secretsmanager.Options{
		BaseEndpoint: aws.String("http://localhost:4566"),
	})
	provider := config.NewCacheDecorator(ctx, config.NewASM("test", client))
	testConfig(t, ctx, provider)
}
