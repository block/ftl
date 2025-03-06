package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/block/scaffolder"
	"github.com/block/scaffolder/extensions/javascript"
	"github.com/radovskyb/watcher"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/log"
)

type schemaGenerateCmd struct {
	Watch          time.Duration `short:"w" help:"Watch template directory at this frequency and regenerate on change."`
	Template       string        `arg:"" help:"Template directory to use." type:"existingdir"`
	Dest           string        `arg:"" help:"Destination directory to write files to (will be erased)."`
	ReconnectDelay time.Duration `help:"Delay before attempting to reconnect to FTL." default:"5s"`
}

func (s *schemaGenerateCmd) Run(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	if s.Watch == 0 {
		return s.oneOffGenerate(ctx, client)
	}
	return s.hotReload(ctx, client)
}

func (s *schemaGenerateCmd) oneOffGenerate(ctx context.Context, schemaClient adminpbconnect.AdminServiceClient) error {
	response, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}
	modules, err := slices.MapErr(response.Msg.Schema.Modules, schema.ModuleFromProto)
	if err != nil {
		return fmt.Errorf("invalid module schema: %w", err)
	}
	return s.regenerateModules(log.FromContext(ctx), modules)
}

func (s *schemaGenerateCmd) hotReload(ctx context.Context, client adminpbconnect.AdminServiceClient) error {
	watch := watcher.New()
	defer watch.Close()

	absTemplatePath, err := filepath.Abs(s.Template)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for template: %w", err)
	}
	absDestPath, err := filepath.Abs(s.Dest)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for destination: %w", err)
	}

	if strings.HasPrefix(absDestPath, absTemplatePath) {
		return fmt.Errorf("destination directory %s must not be inside the template directory %s", absDestPath, absTemplatePath)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Watching %s", s.Template)

	if err := watch.AddRecursive(s.Template); err != nil {
		return fmt.Errorf("failed to watch template directory: %w", err)
	}

	wg, ctx := errgroup.WithContext(ctx)

	moduleChange := make(chan []*schema.Module)

	wg.Go(func() error {
		for {
			stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{SubscriptionId: "cli-schema-generate"}))
			if err != nil {
				return fmt.Errorf("failed to pull schema: %w", err)
			}

			modules := map[string]*schema.Module{}
			for stream.Receive() {
				msg := stream.Msg()
				switch msg := msg.Event.Value.(type) {
				case *schemapb.Notification_FullSchemaNotification:
					for _, m := range msg.FullSchemaNotification.Schema.Modules {
						module, err := schema.ValidatedModuleFromProto(m)
						if err != nil {
							return fmt.Errorf("invalid module: %w", err)
						}
						modules[module.Name] = module
					}
				case *schemapb.Notification_DeploymentRuntimeNotification:
					not, err := schema.DeploymentRuntimeNotificationFromProto(msg.DeploymentRuntimeNotification)
					if err != nil {
						return fmt.Errorf("invalid deployment runtime notification: %w", err)
					}
					if not.Changeset == nil || not.Changeset.IsZero() {
						m, ok := modules[not.Payload.Deployment.Payload.Module]
						if ok {
							err := not.Payload.ApplyToModule(m)
							if err != nil {
								return fmt.Errorf("failed to apply deployment runtime: %w", err)
							}
						}
					}
				case *schemapb.Notification_ChangesetCreatedNotification:
				case *schemapb.Notification_ChangesetPreparedNotification:
				case *schemapb.Notification_ChangesetCommittedNotification:
					for _, m := range msg.ChangesetCommittedNotification.Changeset.RemovingModules {
						delete(modules, m.Name)
					}
					for _, m := range msg.ChangesetCommittedNotification.Changeset.Modules {
						module, err := schema.ValidatedModuleFromProto(m)
						if err != nil {
							return fmt.Errorf("invalid module: %w", err)
						}
						modules[module.Name] = module
					}

					moduleChange <- maps.Values(modules)
				}

				stream.Close()
				logger.Debugf("Stream disconnected, attempting to reconnect...")
				time.Sleep(s.ReconnectDelay)
			}
		}
	})

	wg.Go(func() error { return watch.Start(s.Watch) })

	var previousModules []*schema.Module
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %w", wg.Wait())

		case event := <-watch.Event:
			logger.Debugf("Template changed (%s), regenerating modules", event.Path)
			if err := s.regenerateModules(logger, previousModules); err != nil {
				return fmt.Errorf("failed to regenerate modules: %w", err)
			}

		case modules := <-moduleChange:
			previousModules = modules
			if err := s.regenerateModules(logger, modules); err != nil {
				return fmt.Errorf("failed to regenerate modules: %w", err)
			}
		}
	}
}

func (s *schemaGenerateCmd) regenerateModules(logger *log.Logger, modules []*schema.Module) error {
	if err := os.RemoveAll(s.Dest); err != nil {
		return fmt.Errorf("failed to remove destination directory: %w", err)
	}

	for _, module := range modules {
		if err := scaffolder.Scaffold(s.Template, s.Dest, module,
			scaffolder.Extend(javascript.Extension("template.js", javascript.WithLogger(makeJSLoggerAdapter(logger)))),
		); err != nil {
			return fmt.Errorf("failed to scaffold module %s: %w", module.Name, err)
		}
	}
	logger.Debugf("Generated %d modules in %s", len(modules), s.Dest)
	return nil
}

func makeJSLoggerAdapter(logger *log.Logger) func(args ...any) {
	return func(args ...any) {
		strs := slices.Map(args, func(v any) string { return fmt.Sprintf("%v", v) })
		level := log.Debug
		if prefix, ok := args[0].(string); ok {
			switch prefix {
			case "log:":
				level = log.Info
			case "debug:":
				level = log.Debug
			case "error:":
				level = log.Error
			case "warn:":
				level = log.Warn
			}
		}
		logger.Log(log.Entry{
			Level:   level,
			Message: strings.Join(strs[1:], " "),
		})
	}
}
