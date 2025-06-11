package main

import (
	"context"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

type deployCmd struct {
	Replicas optional.Option[int32] `short:"n" help:"Number of replicas to deploy."`
	NoWait   bool                   `help:"Do not wait for deployment to complete." default:"false"`
	Build    buildCmd               `embed:""`
	Timeout  time.Duration          `short:"t" help:"Timeout for the deployment."`
}

func (d *deployCmd) Run(
	ctx context.Context,
	projConfig projectconfig.Config,
	adminClient adminpbconnect.AdminServiceClient,
	schemaSource *schemaeventsource.EventSource,
) error {
	// logger := log.FromContext(ctx)
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
		return errors.WithStack(errors.New("no directories specified"))
	}

	if !schemaSource.WaitForInitialSync(ctx) {
		return errors.Errorf("timed out waiting for schema sync from server")
	}

	if d.Timeout > 0 {
		var cancel context.CancelFunc //nolint: forbidigo
		ctx, cancel = context.WithTimeoutCause(ctx, d.Timeout, errors.Errorf("terminating deploy due to timeout of %s", d.Timeout))
		defer cancel()
	} else {
		var cancel context.CancelCauseFunc
		ctx, cancel = context.WithCancelCause(ctx)
		defer cancel(errors.Wrap(context.Canceled, "stopping deploy"))
	}

	engine, err := buildengine.New(
		ctx,
		adminClient,
		projConfig,
		d.Build.Dirs,
		false,
		buildengine.BuildEnv(d.Build.BuildEnv),
		buildengine.Parallelism(d.Build.Parallelism),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	// if len(engine.Modules()) == 0 {
	// 	logger.Warnf("No modules were found to build")
	// 	return nil
	// }
	if err := engine.BuildAndDeploy(ctx, schemaSource, d.Replicas); err != nil {
		return errors.Wrap(err, "build failed")
	}
	return nil
}
