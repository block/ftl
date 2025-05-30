package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/buildengine"
	configuration "github.com/block/ftl/internal/config"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
)

const maxLogs = 10

type devCmd struct {
	Watch    time.Duration     `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe  bool              `help:"Do not start the FTL server." default:"false"`
	ServeCmd serveCommonConfig `embed:""`
	Build    buildCmd          `embed:""`
}

func (d *devCmd) Run(
	ctx context.Context,
	cm configuration.Provider[configuration.Configuration],
	sm configuration.Provider[configuration.Secrets],
	projConfig projectconfig.Config,
	bindContext KongContextBinder,
	csm *currentStatusManager,
	cli *SharedCLI,
) error {
	cli.AdminEndpoint = d.ServeCmd.Bind
	os.Unsetenv("FTL_ENDPOINT") //nolint:errcheck
	adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, d.ServeCmd.Bind.String(), log.Error)

	startTime := time.Now()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(errors.Wrap(context.Canceled, "stopping dev server"))
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
		return errors.WithStack(errors.New("no directories specified"))
	}

	// Setup file based logging, which logs everything to a file
	log.SetupDebugFileLogging(ctx, projConfig.Root(), maxLogs)

	// Reset Goose context to ensure it doesn't have any stale state from previous runs.
	if err := resetGooseSession(projConfig); err != nil {
		return errors.Wrap(err, "failed to reset Goose context")
	}

	// Set environment variable so child processes know where modules are expected to be
	// such as via the interactive console or goose.
	absDirs := make([]string, len(d.Build.Dirs))
	for i, dir := range d.Build.Dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return errors.Wrapf(err, "could not create absolute path for %q", dir)
		}
		absDirs[i] = absDir
	}
	os.Setenv("FTL_DEV_DIRS", strings.Join(absDirs, ","))

	source := schemaeventsource.New(ctx, "dev", adminClient)
	terminal.LaunchEmbeddedConsole(ctx, createKongApplication(&DevModeCLI{}, csm), source, func(ctx context.Context, k *kong.Kong, args []string, additionalExit func(int)) error {
		return errors.WithStack(runInnerCmd(ctx, k, projConfig, bindContext, args, additionalExit))
	})
	var deployClient buildengine.AdminClient = adminClient

	g, ctx := errgroup.WithContext(ctx)

	if d.NoServe && d.ServeCmd.Stop {
		logger := log.FromContext(ctx)
		return errors.WithStack(KillBackgroundServe(logger))
	}

	statusManager := terminal.FromContext(ctx)
	defer statusManager.Close()
	starting := statusManager.NewStatus("\u001B[92mStarting FTL Server 🚀\u001B[39m")

	bindAllocator, err := bind.NewBindAllocator(cli.AdminEndpoint, 2)
	if err != nil {
		return errors.Wrap(err, "could not create bind allocator")
	}

	// Default to allowing all origins and headers for console requests in local dev mode.
	d.ServeCmd.Ingress.AllowOrigins = []string{"*"}
	d.ServeCmd.Ingress.AllowHeaders = []string{"*"}

	devModeEndpointUpdates := make(chan dev.LocalEndpoint, 1)

	opts := []buildengine.Option{buildengine.Parallelism(d.Build.Parallelism), buildengine.BuildEnv(d.Build.BuildEnv), buildengine.WithDevMode(devModeEndpointUpdates), buildengine.WithStartTime(startTime)}
	engine, err := buildengine.New(ctx, deployClient, source, projConfig, d.Build.Dirs, false, opts...)
	if err != nil {
		return errors.WithStack(err)
	}

	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)
	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, devModeEndpointUpdates, []rpc.Service{engine})
			if err != nil {
				return errors.Wrap(err, "failed to stop server")
			}
			d.ServeCmd.Stop = false
		}

		g.Go(func() error {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, devModeEndpointUpdates, []rpc.Service{engine})
			if err != nil {
				cancel(errors.Wrap(errors.Join(err, context.Canceled), "dev server failed"))
			} else {
				cancel(errors.Wrap(context.Canceled, "dev server stopped"))
			}
			return errors.WithStack(err)
		})
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}
		starting.Close()

		return errors.WithStack(engine.Dev(ctx, d.Watch))
	})

	err = g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return errors.Wrap(err, "error during dev")
	}
	return nil
}
