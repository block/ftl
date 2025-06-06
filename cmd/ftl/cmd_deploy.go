package main

import (
	"context"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
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
	logger := log.FromContext(ctx)

	if !schemaSource.WaitForInitialSync(ctx) {
		return errors.Errorf("timed out waiting for schema sync from server")
	}
	// Cancel build engine context to ensure all language plugins are killed.
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
		ctx, adminClient, schemaSource, projConfig, d.Build.Dirs, true,
		buildengine.BuildEnv(d.Build.BuildEnv),
		buildengine.Parallelism(d.Build.Parallelism),
	)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(engine.Modules()) == 0 {
		logger.Warnf("No modules were found to deploy")
		return nil
	}
	err = engine.BuildAndDeploy(ctx, d.Replicas, !d.NoWait, true)
	if err != nil {
		return errors.Wrap(err, "failed to deploy")
	}
	logger.Infof("Deployed modules %v", engine.Modules()) //nolint
	terminal.FromContext(ctx).Close()
	return nil
}
