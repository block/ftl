package main

import (
	"context"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/terminal"
	"github.com/block/ftl/internal/test"
)

type testCmd struct {
	ServeCmd serveCommonConfig `embed:""`
}

func (d *testCmd) Run(
	ctx context.Context,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
) error {
	cli.AdminEndpoint = d.ServeCmd.Bind
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "stopping test server"))

	// Setup file based logging, which logs everything to a file
	log.SetupDebugFileLogging(ctx, projConfig.Root(), maxLogs)

	g, ctx := errgroup.WithContext(ctx)

	devModeEndpointUpdates := make(chan dev.LocalEndpoint, 1)
	service := test.NewTestService(projConfig.Name, devModeEndpointUpdates)

	statusManager := terminal.FromContext(ctx)
	defer statusManager.Close()

	bindAllocator, err := bind.NewBindAllocator(d.ServeCmd.Bind, 2)
	if err != nil {
		return errors.Wrap(err, "could not create bind allocator")
	}

	// Default to allowing all origins and headers for console requests in local test mode.
	d.ServeCmd.Ingress.AllowOrigins = []string{"*"}
	d.ServeCmd.Ingress.AllowHeaders = []string{"*"}
	d.ServeCmd.NoConsole = true
	d.ServeCmd.Recreate = true

	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)

	g.Go(func() error {
		err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, devModeEndpointUpdates, []rpc.Service{service})
		if err != nil {
			cancel(errors.Wrap(errors.Join(err, context.Canceled), "test server failed"))
		} else {
			cancel(errors.Wrap(context.Canceled, "test server stopped"))
		}
		return errors.WithStack(err)
	})

	err = g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return errors.Wrap(err, "error during test")
	}
	return nil
}
