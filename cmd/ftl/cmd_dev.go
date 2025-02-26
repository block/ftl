package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
	"github.com/block/ftl/internal/timelineclient"
)

type devCmd struct {
	Watch    time.Duration     `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe  bool              `help:"Do not start the FTL server." default:"false"`
	ServeCmd serveCommonConfig `embed:""`
	Build    buildCmd          `embed:""`
}

func (d *devCmd) Run(
	ctx context.Context,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
	bindContext terminal.KongContextBinder,
	schemaEventSource *schemaeventsource.EventSource,
	timelineClient *timelineclient.Client,
	adminClient ftlv1connect.AdminServiceClient,
	buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	csm *currentStatusManager,
) error {
	startTime := time.Now()
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(fmt.Errorf("stopping dev server: %w", context.Canceled))
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	// Set environment variable so child processes know where modules are expected to be
	// such as via the interactive console or goose.
	absDirs := make([]string, len(d.Build.Dirs))
	for i, dir := range d.Build.Dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("could not create absolute path for %q: %w", dir, err)
		}
		absDirs[i] = absDir
	}
	os.Setenv("FTL_DEV_DIRS", strings.Join(absDirs, ","))

	terminal.LaunchEmbeddedConsole(ctx, createKongApplication(&DevModeCLI{}, csm), bindContext, schemaEventSource)
	var deployClient buildengine.AdminClient = adminClient

	g, ctx := errgroup.WithContext(ctx)

	if d.NoServe && d.ServeCmd.Stop {
		logger := log.FromContext(ctx)
		return KillBackgroundServe(logger)
	}

	statusManager := terminal.FromContext(ctx)
	defer statusManager.Close()
	starting := statusManager.NewStatus("\u001B[92mStarting FTL Server 🚀\u001B[39m")

	bindAllocator, err := bind.NewBindAllocator(d.ServeCmd.Bind, 2)
	if err != nil {
		return fmt.Errorf("could not create bind allocator: %w", err)
	}

	// Default to allowing all origins and headers for console requests in local dev mode.
	d.ServeCmd.Ingress.AllowOrigins = []*url.URL{{Scheme: "*", Host: "*"}}
	d.ServeCmd.Ingress.AllowHeaders = []string{"*"}

	devModeEndpointUpdates := make(chan dev.LocalEndpoint, 1)
	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)
	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, timelineClient, adminClient, buildEngineClient, true, devModeEndpointUpdates)
			if err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}
			d.ServeCmd.Stop = false
		}

		g.Go(func() error {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, timelineClient, adminClient, buildEngineClient, true, devModeEndpointUpdates)
			if err != nil {
				cancel(fmt.Errorf("dev server failed: %w: %w", context.Canceled, err))
			} else {
				cancel(fmt.Errorf("dev server stopped: %w", context.Canceled))
			}
			return err
		})
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}
		starting.Close()

		opts := []buildengine.Option{buildengine.Parallelism(d.Build.Parallelism), buildengine.BuildEnv(d.Build.BuildEnv), buildengine.WithDevMode(devModeEndpointUpdates), buildengine.WithStartTime(startTime)}
		engine, err := buildengine.New(ctx, deployClient, schemaEventSource, projConfig, d.Build.Dirs, d.Build.UpdatesEndpoint, opts...)
		if err != nil {
			return err
		}

		return engine.Dev(ctx, d.Watch)
	})

	err = g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("error during dev: %w", err)
	}
	return nil
}
