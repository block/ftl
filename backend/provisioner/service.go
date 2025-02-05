package provisioner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/BurntSushi/toml"
	"github.com/alecthomas/kong"
	"github.com/puzpuzpuz/xsync/v3"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemaconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/schema/schemaeventsource"
)

// CommonProvisionerConfig is shared config between the production controller and development server.
type CommonProvisionerConfig struct {
	PluginConfigFile *os.File `name:"provisioner-plugin-config" help:"Path to the plugin configuration file." env:"FTL_PROVISIONER_PLUGIN_CONFIG_FILE"`
}

type Config struct {
	Bind               *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8893" env:"FTL_BIND"`
	ControllerEndpoint *url.URL `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	SchemaEndpoint     *url.URL `help:"Schema service endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8897"`
	CommonProvisionerConfig
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c); err != nil {
		panic(err)
	}
}

type Service struct {
	currentModules *xsync.MapOf[string, *schema.Module]
	registry       *ProvisionerRegistry
	eventSource    *schemaeventsource.EventSource
	schemaClient   schemaconnect.SchemaServiceClient
}

func New(
	ctx context.Context,
	config Config,
	registry *ProvisionerRegistry,
	schemaClient schemaconnect.SchemaServiceClient,
) (*Service, error) {

	eventSource := schemaeventsource.New(ctx, "provisioner", schemaClient)
	return &Service{
		currentModules: xsync.NewMapOf[string, *schema.Module](),
		registry:       registry,
		eventSource:    &eventSource,
		schemaClient:   schemaClient,
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

// Start the Provisioner. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	config Config,
	registry *ProvisionerRegistry,
	schemaClient schemaconnect.SchemaServiceClient,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx)
	logger.Debugf("Starting FTL provisioner")

	svc, err := New(ctx, config, registry, schemaClient)
	if err != nil {
		return err
	}
	logger.Debugf("Provisioner available at: %s", config.Bind)
	logger.Debugf("Using FTL endpoint: %s", config.ControllerEndpoint)
	// Hack: as we only have one changeset at a time at the moment, we can just keep track of the last key

	var lastKey key.Changeset
	for event := range channels.IterContext(ctx, svc.eventSource.Events()) {
		if cs, ok := event.ActiveChangeset().Get(); ok {
			if cs.Key != lastKey {
				lastKey = cs.Key
				err := svc.ProvisionChangeset(ctx, cs)
				if err != nil {
					logger.Errorf(err, "Error provisioning changeset")
					continue
				}
				logger.Debugf("Changeset %s provisioned", cs.Key)
			}
		}
	}
	return nil
}

func RegistryFromConfigFile(ctx context.Context, file *os.File, scaling scaling.RunnerScaling) (*ProvisionerRegistry, error) {
	config := provisionerPluginConfig{}
	bytes, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("error reading plugin configuration from %s: %w", file.Name(), err)
	}
	if err := toml.Unmarshal(bytes, &config); err != nil {
		return nil, fmt.Errorf("error parsing plugin configuration: %w", err)
	}

	registry, err := registryFromConfig(ctx, &config, scaling)
	if err != nil {
		return nil, fmt.Errorf("error creating provisioner registry: %w", err)
	}

	return registry, nil
}

func (s *Service) ProvisionChangeset(ctx context.Context, req *schema.Changeset) error {
	logger := log.FromContext(ctx)
	// TODO: Block deployments to make sure only one module is modified at a time
	for _, module := range req.Modules {
		moduleName := module.Name

		existingModule, _ := s.currentModules.Load(moduleName)

		if existingModule != nil {
			syncExistingRuntimes(existingModule, module)
		}

		deployment := s.registry.CreateDeployment(ctx, req.Key, module, existingModule, func(event *schemapb.Event) error {
			_, err := s.schemaClient.UpdateSchema(ctx, connect.NewRequest(&ftlv1.UpdateSchemaRequest{Event: event}))
			if err != nil {
				return fmt.Errorf("error updating schema: %w", err)
			}
			return nil
		})
		running := true
		logger.Debugf("Running deployment for module %s", moduleName)
		for running {
			r, err := deployment.Progress(ctx)
			if err != nil {
				// TODO: Deal with failed deployments
				return fmt.Errorf("error running a provisioner: %w", err)
			}
			running = r
		}

		logger.Debugf("Finished deployment for module %s", moduleName)

	}

	_, err := s.schemaClient.PrepareChangeset(ctx, connect.NewRequest(&ftlv1.PrepareChangesetRequest{Changeset: req.Key.String()}))
	if err != nil {
		return fmt.Errorf("error preparing changeset: %w", err)
	}
	_, err = s.schemaClient.CommitChangeset(ctx, connect.NewRequest(&ftlv1.CommitChangesetRequest{Changeset: req.Key.String()}))
	if err != nil {
		return fmt.Errorf("error committing changeset: %w", err)
	}
	return nil
}

func syncExistingRuntimes(existingModule, desiredModule *schema.Module) {
	existingResources := schema.GetProvisioned(existingModule)
	desiredResources := schema.GetProvisioned(desiredModule)

	for id, desired := range desiredResources {
		if existing, ok := existingResources[id]; ok {
			switch desired := desired.(type) {
			case *schema.Database:
				if existing, ok := existing.(*schema.Database); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Topic:
				if existing, ok := existing.(*schema.Topic); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Verb:
				if existing, ok := existing.(*schema.Verb); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			case *schema.Module:
				if existing, ok := existing.(*schema.Module); ok {
					desired.Runtime = reflect.DeepCopy(existing.Runtime)
				}
			}
		}
	}
}
