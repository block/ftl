package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	osExec "os/exec" //nolint:depguard
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/console"
	"github.com/block/ftl/backend/cron"
	"github.com/block/ftl/backend/ingress"
	"github.com/block/ftl/backend/lease"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/backend/provisioner/scaling/localscaling"
	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	consolefrontend "github.com/block/ftl/frontend/console"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/oci"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type serveCmd struct {
	serveCommonConfig
}

type serveCommonConfig struct {
	IngressBind         *url.URL             `help:"HTTP Ingress bind" default:"http://127.0.0.1:8891"`
	Bind                *url.URL             `help:"Starting endpoint to bind to and advertise to for all FTL services" default:"http://127.0.0.1:8892"`
	DBPort              int                  `help:"Port to use for the database." env:"FTL_DB_PORT" default:"15432"`
	MysqlPort           int                  `help:"Port to use for the MySQL database, if one is required." env:"FTL_MYSQL_PORT" default:"13306"`
	RegistryPort        int                  `help:"Port to use for the registry." env:"FTL_OCI_REGISTRY_PORT" default:"15000"`
	RegistryConfig      oci.RegistryConfig   `prefix:"oci-" hidden:"" embed:""`
	Background          bool                 `help:"Run in the background." default:"false"`
	Stop                bool                 `help:"Stop the running FTL instance. Can be used with --background to restart the server" default:"false"`
	StartupTimeout      time.Duration        `help:"Timeout for the server to start up." default:"10s" env:"FTL_STARTUP_TIMEOUT"`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	DatabaseImage       string               `help:"The container image to start for the database" default:"postgres:15.10" env:"FTL_DATABASE_IMAGE" hidden:""`
	RegistryImage       string               `help:"The container image to start for the image registry" default:"registry:2" env:"FTL_REGISTRY_IMAGE" hidden:""`
	GrafanaImage        string               `help:"The container image to start for the automatic Grafana instance" default:"grafana/otel-lgtm" env:"FTL_GRAFANA_IMAGE" hidden:""`
	EnableGrafana       bool                 `help:"Enable Grafana to view telemetry data." default:"false"`
	NoConsole           bool                 `help:"Disable the console."`
	Ingress             ingress.Config       `embed:"" prefix:"ingress-"`
	Timeline            timeline.Config      `embed:"" prefix:"timeline-"`
	Console             console.Config       `embed:"" prefix:"console-"`
	Admin               admin.Config         `embed:"" prefix:"admin-"`
	Recreate            bool                 `help:"Recreate any stateful resources if they already exist." default:"false"`
	WaitFor             []string             `help:"Wait for these modules to be deployed before becoming ready." placeholder:"MODULE"`
	provisioner.CommonProvisionerConfig
	schemaservice.CommonSchemaServiceConfig
}

const ftlRunningErrorMsg = "FTL is already running. Use 'ftl serve --stop' to stop it"

func (s *serveCmd) Run(
	ctx context.Context,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
) error {
	bindAllocator, err := bind.NewBindAllocator(cli.AdminEndpoint, 2)
	if err != nil {
		return errors.Wrap(err, "could not create bind allocator")
	}
	return s.run(ctx, projConfig, cm, sm, optional.None[chan bool](), false, bindAllocator, nil, nil)
}

//nolint:maintidx
func (s *serveCommonConfig) run(
	ctx context.Context,
	projConfig projectconfig.Config,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	initialised optional.Option[chan bool],
	devMode bool,
	bindAllocator *bind.BindAllocator,
	devModeEndpoints <-chan dev.LocalEndpoint,
	additionalServices []rpc.Service,
) error {
	cli.AdminEndpoint = s.Bind
	os.Unsetenv("FTL_ENDPOINT") //nolint:errcheck

	logger := log.FromContext(ctx)
	services := additionalServices

	timelineClient := timelineclient.NewClient(ctx, s.Bind)
	adminClient := rpc.Dial(adminpbconnect.NewAdminServiceClient, s.Bind.String(), log.Error)
	buildEngineClient := rpc.Dial(buildenginepbconnect.NewBuildEngineServiceClient, s.Bind.String(), log.Error)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, s.Bind.String(), log.Error)

	// We must use our own event source here
	// The injected one is connected to the admin client for CLI commands, we need this one to connect directly
	// to the schema service as it is used by the Admin service
	schemaEventSource := schemaeventsource.New(ctx, "serve", schemaClient)
	router := routing.NewVerbRouter(ctx, schemaEventSource, timelineClient)

	if s.Background {
		if s.Stop {
			// allow usage of --background and --stop together to "restart" the background process
			_ = KillBackgroundServe(logger) //nolint:errcheck // ignore error here if the process is not running
		}
		_, err := adminClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
		if err == nil {
			// The controller is already running, bail out.
			return errors.WithStack(errors.New(ftlRunningErrorMsg))
		}
		if err := runInBackground(logger); err != nil {
			return errors.WithStack(err)
		}

		os.Exit(0)
	}

	if s.Stop {
		return errors.WithStack(KillBackgroundServe(logger))
	}
	if err := writePidFile(os.Getpid()); err != nil {
		logger.Errorf(err, "Failed to write pid file")
	}
	_, err := adminClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err == nil {
		// The controller is already running, bail out.
		return errors.WithStack(errors.New(ftlRunningErrorMsg))
	}

	if s.EnableGrafana && !bool(s.ObservabilityConfig.ExportOTEL) {
		if err := dev.SetupGrafana(ctx, s.GrafanaImage); err != nil {
			logger.Errorf(err, "Failed to setup grafana image")
		} else {
			logger.Infof("Grafana started at http://localhost:3000")
			os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
			os.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "1000")
			s.ObservabilityConfig.ExportOTEL = true
		}
	}
	if err := observability.Init(ctx, false, "", "ftl-serve", ftl.Version, s.ObservabilityConfig); err != nil {
		return errors.Wrap(err, "observability init failed")
	}
	// Bring up the image registry we use to store deployment content
	if err := dev.SetupRegistry(ctx, s.RegistryImage, s.RegistryPort); err != nil {
		return errors.Wrap(err, "registry init failed")
	}
	rc := s.RegistryConfig
	if rc.Registry == "" {
		rc = oci.RegistryConfig{
			AllowInsecure: true,
			Registry:      fmt.Sprintf("127.0.0.1:%d/ftl", s.RegistryPort),
		}
	}
	storage, err := oci.NewArtefactService(ctx, rc)
	if err != nil {
		return errors.Wrap(err, "failed to create OCI registry storage")
	}

	wg, ctx := errgroup.WithContext(ctx)

	if _, err := bindAllocator.Next(); err != nil { // skip the first port, which is used by ingress
		return errors.Wrap(err, "could not allocate port for ingress")
	}
	if _, err := bindAllocator.Next(); err != nil {
		return errors.Wrap(err, "could not allocate port for controller")
	}

	runnerScaling, err := localscaling.NewLocalScaling(
		ctx,
		s.Bind,
		s.Bind,
		s.Bind,
		projConfig.Path,
		!projConfig.DisableIDEIntegration && !projConfig.DisableVSCodeIntegration,
		!projConfig.DisableIDEIntegration && !projConfig.DisableIntellijIntegration,
		storage,
		bool(s.ObservabilityConfig.ExportOTEL),
		devModeEndpoints,
		routing.New(ctx, schemaEventSource),
		schemaClient,
		adminClient,
	)
	if err != nil {
		return errors.WithStack(err)
	}
	err = runnerScaling.Start(ctx)
	if err != nil {
		return errors.Wrap(err, "runner scaling failed to start")
	}

	schemaCtx := log.ContextWithLogger(ctx, logger.Scope("schemaservice"))
	schemaService := schemaservice.NewLocalService(schemaCtx, schemaservice.Config{
		CommonSchemaServiceConfig: s.CommonSchemaServiceConfig,
	}, timelineClient, devMode)
	services = append(services, schemaService)

	if !s.NoConsole {
		svc := console.New(schemaEventSource, timelineClient, adminClient, router, buildEngineClient, s.Bind, s.Console, optional.Some(projConfig), true)
		services = append(services, svc)
		wg.Go(func() error {
			ctx := log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("console"))
			if err := consolefrontend.PrepareServer(ctx); err != nil {
				return errors.Wrap(err, "failed to prepare console server")
			}
			return nil
		})
	}

	provisionerCtx := log.ContextWithLogger(ctx, logger.Scope("provisioner"))

	// default local dev provisioner

	provisionerRegistry := &provisioner.ProvisionerRegistry{
		Bindings: []*provisioner.ProvisionerBinding{
			{
				Provisioner: provisioner.NewDevProvisioner(s.DBPort, s.MysqlPort, s.Recreate),
				Types: []schema.ResourceType{
					schema.ResourceTypeMysql,
					schema.ResourceTypePostgres,
					schema.ResourceTypeTopic,
					schema.ResourceTypeSubscription,
					schema.ResourceTypeImage,
				},
				ID: "dev",
			},
			{
				Provisioner: provisioner.NewEgressProvisioner(adminClient),
				Types: []schema.ResourceType{
					schema.ResourceTypeEgress,
				},
				ID: "egress",
			},
			{
				Provisioner: provisioner.NewSQLMigrationProvisioner(storage),
				Types:       []schema.ResourceType{schema.ResourceTypeSQLMigration},
				ID:          "migration",
			},
			{
				Provisioner: provisioner.NewRunnerScalingProvisioner(runnerScaling, true),
				Types:       []schema.ResourceType{schema.ResourceTypeRunner},
				ID:          "runner",
			},
			{
				Provisioner: provisioner.NewFixtureProvisioner(),
				Types:       []schema.ResourceType{schema.ResourceTypeFixture},
				ID:          "fixture",
			},
		},
	}

	// read provisioners from a config file if provided
	if s.PluginConfigFile != nil {
		r, err := provisioner.RegistryFromConfigFile(provisionerCtx, s.WorkingDir, s.PluginConfigFile, runnerScaling, adminClient, storage)
		if err != nil {
			return errors.Wrap(err, "failed to create provisioner registry")
		}
		provisionerRegistry = r
	}

	wg.Go(func() error {
		if err := provisioner.Start(provisionerCtx, provisionerRegistry, schemaClient, timelineClient); err != nil {
			logger.Errorf(err, "provisionerfailed: %v", err)
			return errors.Wrap(err, "provisionerfailed")
		}
		return nil
	})

	// Start Timeline
	timelineService, err := timeline.New(ctx, s.Timeline)
	if err != nil {
		return errors.Wrap(err, "failed to create timeline service")
	}
	services = append(services, timelineService)
	// Start Cron
	wg.Go(func() error {
		ctx := log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("cron"))
		c := cron.Config{
			SchemaServiceEndpoint: s.Bind,
			TimelineEndpoint:      s.Bind,
		}
		err := cron.Start(ctx, c, schemaEventSource, router, timelineClient)
		if err != nil {
			return errors.Wrap(err, "cron failed")
		}
		return nil
	})
	// Start Ingress
	wg.Go(func() error {
		ctx := log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("http-ingress"))
		err := ingress.Start(ctx, s.IngressBind, s.Ingress, schemaEventSource, router, timelineClient)
		if err != nil {
			return errors.Wrap(err, "ingress failed")
		}
		return nil
	})
	services = append(services, lease.New(ctx))
	// Start Admin
	adminService := admin.NewAdminService(s.Admin, cm, sm, schemaClient, schemaEventSource, storage, router, timelineClient, s.WaitFor)
	services = append(services, adminService)

	// Start the common server
	wg.Go(func() error {
		err := rpc.Serve(ctx, s.Bind, rpc.WithServices(services...))
		if err != nil {
			return errors.Wrap(err, "lease failed")
		}
		return nil
	})
	// Wait for controller to start, then run startup commands.
	wg.Go(func() error {
		start := time.Now()

		err := rpc.Wait(ctx, backoff.Backoff{Min: time.Millisecond * 10, Max: time.Millisecond * 50}, time.Minute*100, schemaClient)
		if err != nil {
			logger.Errorf(err, "FTL failed to start")
		}
		logger.Infof("FTL started in %.2fs", time.Since(start).Seconds()) //nolint

		if len(projConfig.Commands.Startup) > 0 {
			for _, cmd := range projConfig.Commands.Startup {
				logger.Debugf("Executing startup command: %s", cmd)
				if err := exec.Command(ctx, log.Info, ".", "bash", "-c", cmd).Run(); err != nil {
					return errors.Wrap(err, "startup command failed")
				}
			}
		}

		if ch, ok := initialised.Get(); ok {
			ch <- true
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return errors.Wrap(err, "serve failed")
	}

	return nil
}

func runInBackground(logger *log.Logger) error {
	if running, err := isServeRunning(logger); err != nil {
		return errors.Wrap(err, "failed to check if FTL is running")
	} else if running {
		logger.Warnf(ftlRunningErrorMsg)
		return nil
	}

	args := make([]string, 0, len(os.Args))
	for _, arg := range os.Args[1:] {
		if arg == "--background" || arg == "--stop" {
			continue
		}
		args = append(args, arg)
	}

	cmd := osExec.Command(os.Args[0], args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = nil, nil, nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start background process")
	}

	pid := cmd.Process.Pid
	err := writePidFile(pid)
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("`ftl serve` running in background with pid: %d", pid)
	return nil
}

func writePidFile(pid int) error {
	pidFilePath, err := pidFilePath()
	if err != nil {
		return errors.Wrap(err, "failed to get pid file path")
	}
	if err := os.MkdirAll(filepath.Dir(pidFilePath), 0750); err != nil {
		return errors.Wrap(err, "failed to create directory for pid file")
	}

	if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(pid)), 0600); err != nil {
		return errors.Wrap(err, "failed to write pid file")
	}
	return nil
}

func KillBackgroundServe(logger *log.Logger) error {
	pidFilePath, err := pidFilePath()
	if err != nil {
		logger.Infof("No background process found")
		return errors.WithStack(err)
	}

	pid, err := getPIDFromPath(pidFilePath)
	if err != nil || pid == 0 {
		logger.Debugf("FTL serve is not running in the background")
		return nil
	}

	if err := os.Remove(pidFilePath); err != nil {
		logger.Errorf(err, "Failed to remove pid file: %v", err)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if !errors.Is(err, syscall.ESRCH) {
			return errors.WithStack(err)
		}
	}

	logger.Infof("`ftl serve` stopped (pid: %d)", pid)
	return nil
}

func pidFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.WithStack(err)
	}
	return filepath.Join(homeDir, ".ftl", "ftl-serve.pid"), nil
}

func getPIDFromPath(path string) (int, error) {
	pidBytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0, nil
	} else if err != nil {
		return 0, errors.WithStack(err)
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return pid, nil
}

func isServeRunning(logger *log.Logger) (bool, error) {
	pidFilePath, err := pidFilePath()
	if err != nil {
		return false, errors.WithStack(err)
	}

	pid, err := getPIDFromPath(pidFilePath)
	if err != nil || pid == 0 {
		return false, errors.WithStack(err)
	}

	err = syscall.Kill(pid, 0)
	if err != nil {
		if errors.Is(err, syscall.ESRCH) {
			logger.Infof("Process with PID %d does not exist.", pid)
			return false, nil
		}
		if errors.Is(err, syscall.EPERM) {
			logger.Infof("Process with PID %d exists but no permission to signal it.", pid)
			return true, nil
		}
		return false, errors.WithStack(err)
	}

	return true, nil
}
