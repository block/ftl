package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	osExec "os/exec" //nolint:depguard
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl"
	"github.com/block/ftl/backend/admin"
	"github.com/block/ftl/backend/console"
	"github.com/block/ftl/backend/controller"
	"github.com/block/ftl/backend/controller/artefacts"
	"github.com/block/ftl/backend/cron"
	"github.com/block/ftl/backend/ingress"
	"github.com/block/ftl/backend/lease"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/buildengine/v1/buildenginepbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/backend/provisioner/scaling/localscaling"
	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/schema"
	consolefrontend "github.com/block/ftl/frontend/console"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/exec"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/routing"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/timelineclient"
)

type serveCmd struct {
	serveCommonConfig
}

type serveCommonConfig struct {
	Bind                *url.URL             `help:"Starting endpoint to bind to and advertise to. Each controller, ingress, runner and language plugin will increment the port by 1" default:"http://127.0.0.1:8891"`
	DBPort              int                  `help:"Port to use for the database." env:"FTL_DB_PORT" default:"15432"`
	MysqlPort           int                  `help:"Port to use for the MySQL database, if one is required." env:"FTL_MYSQL_PORT" default:"13306"`
	RegistryPort        int                  `help:"Port to use for the registry." env:"FTL_OCI_REGISTRY_PORT" default:"15000"`
	Controllers         int                  `short:"c" help:"Number of controllers to start." default:"1"`
	Provisioners        int                  `short:"p" help:"Number of provisioners to start." default:"1"`
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
	Lease               lease.Config         `embed:"" prefix:"lease-"`
	Admin               admin.Config         `embed:"" prefix:"admin-"`
	Recreate            bool                 `help:"Recreate any stateful resources if they already exist." default:"false"`
	controller.CommonConfig
	provisioner.CommonProvisionerConfig
	schemaservice.CommonSchemaServiceConfig
}

const ftlRunningErrorMsg = "FTL is already running. Use 'ftl serve --stop' to stop it"

func (s *serveCmd) Run(
	ctx context.Context,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
	controllerClient ftlv1connect.ControllerServiceClient,
	timelineClient *timelineclient.Client,
	adminClient admin.Client,
	schemaClient ftlv1connect.SchemaServiceClient,
	schemaEventSourceFactory func() schemaeventsource.EventSource,
	buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
) error {
	bindAllocator, err := bind.NewBindAllocator(s.Bind, 2)
	if err != nil {
		return fmt.Errorf("could not create bind allocator: %w", err)
	}
	return s.run(ctx, projConfig, cm, sm, optional.None[chan bool](), false, bindAllocator, controllerClient, timelineClient, adminClient, schemaClient, schemaEventSourceFactory, buildEngineClient, s.Recreate, nil)
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
	controllerClient ftlv1connect.ControllerServiceClient,
	timelineClient *timelineclient.Client,
	adminClient admin.Client,
	schemaClient ftlv1connect.SchemaServiceClient,
	schemaEventSourceFactory func() schemaeventsource.EventSource,
	buildEngineClient buildenginepbconnect.BuildEngineServiceClient,
	recreate bool,
	devModeEndpoints <-chan dev.LocalEndpoint,
) error {

	logger := log.FromContext(ctx)

	if s.Background {
		if s.Stop {
			// allow usage of --background and --stop together to "restart" the background process
			_ = KillBackgroundServe(logger) //nolint:errcheck // ignore error here if the process is not running
		}
		_, err := controllerClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
		if err == nil {
			// The controller is already running, bail out.
			return errors.New(ftlRunningErrorMsg)
		}
		if err := runInBackground(logger); err != nil {
			return err
		}

		if err := waitForControllerOnline(ctx, s.StartupTimeout, controllerClient); err != nil {
			return err
		}

		os.Exit(0)
	}

	if s.Stop {
		return KillBackgroundServe(logger)
	}
	_, err := controllerClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
	if err == nil {
		// The controller is already running, bail out.
		return errors.New(ftlRunningErrorMsg)
	}
	if s.Provisioners > 0 {
		logger.Debugf("Starting FTL with %d controller(s) and %d provisioner(s)", s.Controllers, s.Provisioners)
	} else {
		logger.Debugf("Starting FTL with %d controller(s)", s.Controllers)
	}

	if s.EnableGrafana && !bool(s.ObservabilityConfig.ExportOTEL) {
		err := dev.SetupGrafana(ctx, s.GrafanaImage)
		if err != nil {
			logger.Errorf(err, "Failed to setup grafana image")
		} else {
			logger.Infof("Grafana started at http://localhost:3000")
			os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
			os.Setenv("OTEL_METRIC_EXPORT_INTERVAL", "1000")
			s.ObservabilityConfig.ExportOTEL = true
		}
	}
	err = observability.Init(ctx, false, "", "ftl-serve", ftl.Version, s.ObservabilityConfig)
	if err != nil {
		return fmt.Errorf("observability init failed: %w", err)
	}
	// Bring up the image registry we use to store deployment content
	err = dev.SetupRegistry(ctx, s.RegistryImage, s.RegistryPort)
	if err != nil {
		return fmt.Errorf("registry init failed: %w", err)
	}
	storage, err := artefacts.NewOCIRegistryStorage(artefacts.RegistryConfig{
		AllowInsecure: true,
		Registry:      fmt.Sprintf("127.0.0.1:%d/ftl", s.RegistryPort),
	})
	if err != nil {
		return fmt.Errorf("failed to create OCI registry storage: %w", err)
	}

	wg, ctx := errgroup.WithContext(ctx)

	controllerAddresses := make([]*url.URL, 0, s.Controllers)
	controllerIngressAddresses := make([]*url.URL, 0, s.Controllers)
	for range s.Controllers {
		ingressBind, err := bindAllocator.Next()
		if err != nil {
			return fmt.Errorf("could not allocate port for controller ingress: %w", err)
		}
		controllerIngressAddresses = append(controllerIngressAddresses, ingressBind)
		controllerBind, err := bindAllocator.Next()
		if err != nil {
			return fmt.Errorf("could not allocate port for controller: %w", err)
		}
		controllerAddresses = append(controllerAddresses, controllerBind)
	}

	// Add console addresses to allow origins for console requests
	consoleURLs := []string{
		"http://localhost:8899",
		"http://127.0.0.1:8899",
	}
	for _, urlStr := range consoleURLs {
		consoleURL, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("could not parse console URL %q: %w", urlStr, err)
		}
		s.Ingress.AllowOrigins = append(s.Ingress.AllowOrigins, consoleURL)
	}

	provisionerAddresses := make([]*url.URL, 0, s.Provisioners)
	for range s.Provisioners {
		bind, err := bindAllocator.Next()
		if err != nil {
			return fmt.Errorf("could not allocate port for provisioner: %w", err)
		}
		provisionerAddresses = append(provisionerAddresses, bind)
	}

	schemaBind, err := url.Parse("http://localhost:8897")
	if err != nil {
		return fmt.Errorf("failed to parse bind URL: %w", err)
	}

	runnerScaling, err := localscaling.NewLocalScaling(
		ctx,
		controllerAddresses,
		schemaBind,
		s.Lease.Bind,
		projConfig.Path,
		devMode && !projConfig.DisableIDEIntegration,
		storage,
		bool(s.ObservabilityConfig.ExportOTEL),
		devModeEndpoints,
	)
	if err != nil {
		return err
	}
	err = runnerScaling.Start(ctx)
	if err != nil {
		return fmt.Errorf("runner scaling failed to start: %w", err)
	}

	schemaCtx := log.ContextWithLogger(ctx, logger.Scope("schemaservice"))
	wg.Go(func() error {
		// TODO: Allocate properly, and support multiple instances
		config := schemaservice.Config{
			CommonSchemaServiceConfig: s.CommonSchemaServiceConfig,
			Bind:                      schemaBind,
		}
		if err := schemaservice.Start(schemaCtx, config); err != nil {
			logger.Errorf(err, "schemaservice failed: %v", err)
			return fmt.Errorf("schemaservice failed: %w", err)
		}
		return nil
	})

	for i := range s.Controllers {
		config := controller.Config{
			CommonConfig: s.CommonConfig,
			Bind:         controllerAddresses[i],
			Key:          key.NewLocalControllerKey(i),
		}
		config.SetDefaults()
		config.ModuleUpdateFrequency = time.Second * 1

		scope := fmt.Sprintf("controller%d", i)
		controllerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

		wg.Go(func() error {
			if err := controller.Start(controllerCtx, config, storage, adminClient, timelineClient, schemaClient, true); err != nil {
				logger.Errorf(err, "controller%d failed: %v", i, err)
				return fmt.Errorf("controller%d failed: %w", i, err)
			}
			return nil
		})
	}

	if !s.NoConsole {
		if err := consolefrontend.PrepareServer(ctx); err != nil {
			return fmt.Errorf("failed to prepare console server: %w", err)
		}
		wg.Go(func() error {
			// Deliberately start Console in the foreground.
			ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("console"))
			err := console.Start(ctx, s.Console, schemaEventSourceFactory(), timelineClient, adminClient, routing.NewVerbRouter(ctx, schemaEventSourceFactory(), timelineClient), buildEngineClient)
			if err != nil {
				return fmt.Errorf("failed to start console server: %w", err)
			}
			return nil
		})
	}

	for i := range s.Provisioners {
		config := provisioner.Config{
			Bind:                    provisionerAddresses[i],
			ControllerEndpoint:      controllerAddresses[i%len(controllerAddresses)],
			CommonProvisionerConfig: s.CommonProvisionerConfig,
		}

		config.SetDefaults()

		scope := fmt.Sprintf("provisioner%d", i)
		provisionerCtx := log.ContextWithLogger(ctx, logger.Scope(scope))

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
					},
					ID: "dev",
				},
				{
					Provisioner: provisioner.NewSQLMigrationProvisioner(storage),
					Types:       []schema.ResourceType{schema.ResourceTypeSQLMigration},
					ID:          "migration",
				},
				{
					Provisioner: provisioner.NewRunnerScalingProvisioner(runnerScaling),
					Types:       []schema.ResourceType{schema.ResourceTypeRunner},
					ID:          "runner",
				},
			},
		}

		// read provisioners from a config file if provided
		if s.PluginConfigFile != nil {
			r, err := provisioner.RegistryFromConfigFile(provisionerCtx, s.PluginConfigFile, runnerScaling)
			if err != nil {
				return fmt.Errorf("failed to create provisioner registry: %w", err)
			}
			provisionerRegistry = r
		}

		wg.Go(func() error {
			if err := provisioner.Start(provisionerCtx, config, provisionerRegistry, schemaClient); err != nil {
				logger.Errorf(err, "provisioner%d failed: %v", i, err)
				return fmt.Errorf("provisioner%d failed: %w", i, err)
			}
			return nil
		})
	}

	// Start Timeline
	wg.Go(func() error {
		err := timeline.Start(ctx, s.Timeline)
		if err != nil {
			return fmt.Errorf("timeline failed: %w", err)
		}
		return nil
	})
	// Start Cron
	wg.Go(func() error {
		ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("cron"))
		err := cron.Start(ctx, schemaEventSourceFactory(), routing.NewVerbRouter(ctx, schemaEventSourceFactory(), timelineClient), timelineClient)
		if err != nil {
			return fmt.Errorf("cron failed: %w", err)
		}
		return nil
	})
	// Start Ingress
	wg.Go(func() error {
		ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).Scope("http-ingress"))
		err := ingress.Start(ctx, s.Ingress, schemaClient, routing.NewVerbRouter(ctx, schemaEventSourceFactory(), timelineClient), timelineClient)
		if err != nil {
			return fmt.Errorf("ingress failed: %w", err)
		}
		return nil
	})
	// Start Leases
	wg.Go(func() error {
		err := lease.Start(ctx, s.Lease)
		if err != nil {
			return fmt.Errorf("lease failed: %w", err)
		}
		return nil
	})
	// Start Admin
	wg.Go(func() error {
		err := admin.Start(ctx, s.Admin, cm, sm, admin.NewSchemaRetreiver(schemaEventSourceFactory()))
		if err != nil {
			return fmt.Errorf("lease failed: %w", err)
		}
		return nil
	})
	// Wait for controller to start, then run startup commands.
	wg.Go(func() error {
		start := time.Now()
		if err := waitForControllerOnline(ctx, s.StartupTimeout, controllerClient); err != nil {
			return fmt.Errorf("controller failed to start: %w", err)
		}
		logger.Infof("Controller started in %.2fs", time.Since(start).Seconds())

		if len(projConfig.Commands.Startup) > 0 {
			for _, cmd := range projConfig.Commands.Startup {
				logger.Debugf("Executing startup command: %s", cmd)
				if err := exec.Command(ctx, log.Info, ".", "bash", "-c", cmd).Run(); err != nil {
					return fmt.Errorf("startup command failed: %w", err)
				}
			}
		}

		if ch, ok := initialised.Get(); ok {
			ch <- true
		}
		return nil
	})

	if err := wg.Wait(); err != nil {
		return fmt.Errorf("serve failed: %w", err)
	}

	return nil
}

func runInBackground(logger *log.Logger) error {
	if running, err := isServeRunning(logger); err != nil {
		return fmt.Errorf("failed to check if FTL is running: %w", err)
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
		return fmt.Errorf("failed to start background process: %w", err)
	}

	pidFilePath, err := pidFilePath()
	if err != nil {
		return fmt.Errorf("failed to get pid file path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(pidFilePath), 0750); err != nil {
		return fmt.Errorf("failed to create directory for pid file: %w", err)
	}

	if err := os.WriteFile(pidFilePath, []byte(strconv.Itoa(cmd.Process.Pid)), 0600); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	logger.Infof("`ftl serve` running in background with pid: %d", cmd.Process.Pid)
	return nil
}

func KillBackgroundServe(logger *log.Logger) error {
	pidFilePath, err := pidFilePath()
	if err != nil {
		logger.Infof("No background process found")
		return err
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
			return err
		}
	}

	logger.Infof("`ftl serve` stopped (pid: %d)", pid)
	return nil
}

func pidFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".ftl", "ftl-serve.pid"), nil
}

func getPIDFromPath(path string) (int, error) {
	pidBytes, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func isServeRunning(logger *log.Logger) (bool, error) {
	pidFilePath, err := pidFilePath()
	if err != nil {
		return false, err
	}

	pid, err := getPIDFromPath(pidFilePath)
	if err != nil || pid == 0 {
		return false, err
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
		return false, err
	}

	return true, nil
}

// waitForControllerOnline polls the controller service until it is online.
func waitForControllerOnline(ctx context.Context, startupTimeout time.Duration, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Waiting %s for controller to be online", startupTimeout)

	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	ticker := time.NewTicker(time.Millisecond * 50)
	defer ticker.Stop()

	for range channels.IterContext(ctx, ticker.C) {
		_, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
		if err != nil {
			logger.Tracef("Error getting status, retrying...: %v", err)
			continue // retry
		}

		return nil
	}
	if ctx.Err() == nil {
		return nil
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		logger.Errorf(ctx.Err(), "Timeout reached while polling for controller status")
	}
	return fmt.Errorf("context cancelled: %w", ctx.Err())
}
