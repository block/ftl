package main

import (
	"context"
	"fmt"
	"time"

	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type deployCmd struct {
	Replicas int32         `short:"n" help:"Number of replicas to deploy." default:"1"`
	NoWait   bool          `help:"Do not wait for deployment to complete." default:"false"`
	Build    buildCmd      `embed:""`
	Timeout  time.Duration `short:"t" help:"Timeout for the deployment."`
}

func (d *deployCmd) Run(
	ctx context.Context,
	projConfig projectconfig.Config,
	provisionerClient provisionerconnect.ProvisionerServiceClient,
	schemaServiceClient ftlv1connect.SchemaServiceClient,
	schemaSourceFactory func() schemaeventsource.EventSource,
) error {
	// Cancel build engine context to ensure all language plugins are killed.
	if d.Timeout > 0 {
		var cancel context.CancelFunc //nolint: forbidigo
		ctx, cancel = context.WithTimeoutCause(ctx, d.Timeout, fmt.Errorf("terminating deploy due to timeout of %s", d.Timeout))
		defer cancel()
	} else {
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		defer cancel(fmt.Errorf("stopping deploy"))
	}
	engine, err := buildengine.New(
		ctx, provisionerClient, schemaServiceClient, schemaSourceFactory(), projConfig, d.Build.Dirs, d.Build.UpdatesEndpoint,
		buildengine.BuildEnv(d.Build.BuildEnv),
		buildengine.Parallelism(d.Build.Parallelism),
	)
	if err != nil {
		return err
	}

	err = engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait)
	if err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}
	return nil
}
