package provisioner

import (
	"bufio"
	"context"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/BurntSushi/toml"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemaconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/channels"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	timeline "github.com/block/ftl/internal/timelineclient"
)

// CommonProvisionerConfig is shared config between the production controller and development server.
type CommonProvisionerConfig struct {
	PluginConfigFile *os.File `name:"provisioner-plugin-config" help:"Path to the plugin configuration file." env:"FTL_PROVISIONER_PLUGIN_CONFIG_FILE"`
	WorkingDir       string   `help:"Working directory." env:"FTL_WORKING_DIR" default:"."`
}

type Config struct {
	SchemaEndpoint   *url.URL `help:"Schema service endpoint." env:"FTL_SCHEMA_ENDPOINT" default:"http://127.0.0.1:8892"`
	AdminEndpoint    *url.URL `help:"Admin service endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
	TimelineEndpoint *url.URL `help:"Timeline endpoint." env:"FTL_TIMELINE_ENDPOINT" default:"http://127.0.0.1:8892"`
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
	registry *ProvisionerRegistry,
	schemaClient schemaconnect.SchemaServiceClient,
) (*Service, error) {

	eventSource := schemaeventsource.New(ctx, "provisioner", schemaClient)
	return &Service{
		currentModules: xsync.NewMapOf[string, *schema.Module](),
		registry:       registry,
		eventSource:    eventSource,
		schemaClient:   schemaClient,
	}, nil
}

func (s *Service) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

// Start the Provisioner. Blocks until the context is cancelled.
func Start(
	ctx context.Context,
	registry *ProvisionerRegistry,
	schemaClient schemaconnect.SchemaServiceClient,
	timelineClient *timeline.Client,
) error {
	timelineLogSink := timeline.NewLogSink(timelineClient, log.Debug)
	go timelineLogSink.RunLogLoop(ctx)
	logger := log.FromContext(ctx).AddSink(timelineLogSink)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Debugf("Starting FTL provisioner")

	svc, err := New(ctx, registry, schemaClient)
	if err != nil {
		return errors.WithStack(err)
	}

	for event := range channels.IterContext(ctx, svc.eventSource.Subscribe(ctx)) {
		go func() {
			switch e := event.(type) {
			case *schema.ChangesetCreatedNotification:
				logger.Debugf("Received ChangesetCreatedNotification %s", e.Changeset.Key)
				err := svc.HandleChangesetPreparing(ctx, e.Changeset)
				if err != nil {
					logger.Errorf(err, "Error provisioning changeset")
					_, err = svc.schemaClient.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: e.Changeset.Key.String(), Error: err.Error()}))
					if err != nil {
						logger.Errorf(err, "Error rolling back changeset")
					}
				}
			case *schema.ChangesetPreparedNotification:
				logger.Debugf("Received ChangesetPreparedNotification %s", e.Key)
				err := svc.HandleChangesetPrepared(ctx, e.Key)
				if err != nil {
					_, err := svc.schemaClient.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: e.Key.String(), Error: err.Error()}))
					logger.Errorf(err, "Error provisioning changeset")
				}
			case *schema.ChangesetCommittedNotification:
				logger.Debugf("Received ChangesetCommittedNotification %s", e.Changeset.Key)
				err := svc.HandleChangesetCommitted(ctx, e.Changeset)
				if err != nil {
					logger.Errorf(err, "Error provisioning changeset")
				}
			case *schema.ChangesetDrainedNotification:
				logger.Debugf("Received ChangesetDrainedNotification %s", e.Key)
				err := svc.HandleChangesetDrained(ctx, e.Key)
				if err != nil {
					logger.Errorf(err, "Error de-provisioning changeset")
				}
			case *schema.ChangesetRollingBackNotification:
				logger.Debugf("Received ChangesetRollingBackNotification %s", e.Changeset.Key)
				err := svc.HandleChangesetRollingBack(ctx, e.Changeset)
				if err != nil {
					logger.Errorf(err, "Error de-provisioning changeset")
				}
			case *schema.DeploymentRuntimeNotification:
				logger.Debugf("Received DeploymentRuntimeNotification %s", e.Payload.Deployment)
				if e.Changeset == nil || e.Changeset.IsZero() {
					// This is updated outside of a changeset, so we need to handle it
					switch element := e.Payload.Element.(type) {
					case *schema.ModuleRuntimeScaling:
						err := svc.handleScalingUpdate(ctx, e.Payload.Deployment, element)
						if err != nil {
							logger.Errorf(err, "Error scaling runtime")
						}
					default:

					}
				}
			case *schema.FullSchemaNotification:
				logger.Debugf("Provisioning changesets from full schema notification")
				for _, cs := range e.Changesets {
					if cs.State == schema.ChangesetStatePreparing {
						err := svc.HandleChangesetPreparing(ctx, cs)
						if err != nil {
							logger.Errorf(err, "Error provisioning changeset")
							_, err = svc.schemaClient.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: cs.Key.String(), Error: err.Error()}))
							if err != nil {
								logger.Errorf(err, "Error rolling back changeset")
							}
							continue
						}
					} else if cs.State == schema.ChangesetStatePrepared {
						err := svc.HandleChangesetPrepared(ctx, cs.Key)
						if err != nil {
							logger.Errorf(err, "Error provisioning changeset")
							_, err := svc.schemaClient.RollbackChangeset(ctx, connect.NewRequest(&ftlv1.RollbackChangesetRequest{Changeset: cs.Key.String(), Error: err.Error()}))
							if err != nil {
								logger.Errorf(err, "error rolling back changeset")
							}
							continue
						}
					} else if cs.State == schema.ChangesetStateCommitted {
						err := svc.HandleChangesetCommitted(ctx, cs)
						if err != nil {
							logger.Errorf(err, "Error provisioning changeset")
							continue
						}
					} else if cs.State == schema.ChangesetStateDrained {
						err := svc.HandleChangesetDrained(ctx, cs.Key)
						if err != nil {
							logger.Errorf(err, "Error de-provsisiong changeset")
							continue
						}
					} else if cs.State == schema.ChangesetStateRollingBack {
						err := svc.HandleChangesetRollingBack(ctx, cs)
						if err != nil {
							// We just ignore this error, as failing to roll back could be because we failed to provision
							logger.Errorf(err, "Error rolling back changeset")
						}
					}
				}
			case *schema.ChangesetFailedNotification, *schema.ChangesetFinalizedNotification:
			}

		}()
	}
	return nil
}

func RegistryFromConfigFile(ctx context.Context, workingDir string, file *os.File, scaling scaling.RunnerScaling, adminClient adminpbconnect.AdminServiceClient) (*ProvisionerRegistry, error) {
	config := provisionerPluginConfig{}
	bytes, err := io.ReadAll(bufio.NewReader(file))
	if err != nil {
		return nil, errors.Wrapf(err, "error reading plugin configuration from %s", file.Name())
	}
	if err := toml.Unmarshal(bytes, &config); err != nil {
		return nil, errors.Wrap(err, "error parsing plugin configuration")
	}

	registry, err := registryFromConfig(ctx, workingDir, &config, scaling, adminClient)
	if err != nil {
		return nil, errors.Wrap(err, "error creating provisioner registry")
	}

	return registry, nil
}

func (s *Service) HandleChangesetPrepared(ctx context.Context, req key.Changeset) error {

	_, err := s.schemaClient.CommitChangeset(ctx, connect.NewRequest(&ftlv1.CommitChangesetRequest{Changeset: req.String()}))
	if err != nil {
		return errors.Wrap(err, "error committing changeset")
	}
	return nil
}

func (s *Service) HandleChangesetCommitted(ctx context.Context, req *schema.Changeset) error {
	go func() {
		time.Sleep(time.Second * 5)
		_, err := s.schemaClient.DrainChangeset(ctx, connect.NewRequest(&ftlv1.DrainChangesetRequest{Changeset: req.Key.String()}))
		if err != nil {
			log.FromContext(ctx).Errorf(err, "Error draining changeset")
		}
	}()
	return nil
}

func (s *Service) HandleChangesetDrained(ctx context.Context, cs key.Changeset) error {
	logger := log.FromContext(ctx).Changeset(cs)
	changeset := s.eventSource.ActiveChangesets()[cs]

	moduleNames := slices.Map(changeset.InternalRealm().Modules, func(m *schema.Module) string {
		return m.Name
	})

	err := s.deProvision(ctx, cs, changeset.RealmChanges)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = s.schemaClient.FinalizeChangeset(ctx, connect.NewRequest(&ftlv1.FinalizeChangesetRequest{Changeset: cs.String()}))
	if err != nil {
		return errors.Wrap(err, "error finalizing changeset")
	}
	logger.Debugf("Successfully completed deployment for changeset %s [%s]", cs, strings.Join(moduleNames, ","))
	return nil
}

func (s *Service) HandleChangesetRollingBack(ctx context.Context, changeset *schema.Changeset) error {
	logger := log.FromContext(ctx).Changeset(changeset.Key)
	err := s.deProvision(ctx, changeset.Key, changeset.RealmChanges)
	if err != nil {
		logger.Errorf(err, "Error de-provisioning changeset")
	}
	_, err = s.schemaClient.FailChangeset(ctx, connect.NewRequest(&ftlv1.FailChangesetRequest{Changeset: changeset.Key.String()}))
	if err != nil {
		return errors.Wrap(err, "error finalizing changeset")
	}
	logger.Debugf("Completed rollback for changeset %s", changeset.Key)
	return nil
}

func (s *Service) deProvision(ctx context.Context, cs key.Changeset, realmChanges []*schema.RealmChange) error {

	logger := log.FromContext(ctx)
	group := errgroup.Group{}
	for _, rc := range realmChanges {
		for _, module := range rc.RemovingModules {
			moduleName := module.Name

			group.Go(func() error {
				var current *schema.Module
				existing := s.eventSource.CanonicalView().Module(rc.Name, moduleName)
				if f, ok := existing.Get(); ok {
					current = f
				}
				deployment := s.registry.CreateDeployment(ctx, cs, module, current, func(element *schema.RuntimeElement) error {
					cs := cs.String()
					_, err := s.schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
						Changeset: &cs,
						Update:    element.ToProto(),
					}))
					if err != nil {
						return errors.Wrap(err, "error updating runtime")
					}
					return nil
				})
				if err := deployment.Run(ctx); err != nil {
					return errors.Wrap(err, "error running deployment")
				}
				logger.Debugf("Finished de-provisioning for module %s", moduleName)
				return nil
			})

		}
	}
	err := group.Wait()
	if err != nil {
		return errors.Wrap(err, "error running deployments")
	}
	return nil
}

func (s *Service) HandleChangesetPreparing(ctx context.Context, req *schema.Changeset) error {
	mLogger := log.FromContext(ctx).Changeset(req.Key)
	moduleNames := slices.Map(req.InternalRealm().Modules, func(m *schema.Module) string {
		return m.Name
	})
	mLogger.Debugf("Starting deployment for changeset %s [%s]", req.Key, strings.Join(moduleNames, ","))
	group := errgroup.Group{}
	// TODO: Block deployments to make sure only one module is modified at a time
	for _, module := range req.InternalRealm().Modules {
		logger := mLogger.Module(module.Name)
		ctx := log.ContextWithLogger(ctx, logger)
		moduleName := module.Name

		existingModule, _ := s.currentModules.Load(moduleName)

		if existingModule != nil {
			syncExistingRuntimes(existingModule, module)
		}
		group.Go(func() error {
			if err := s.registry.VerifyDeploymentSupported(ctx, module); err != nil {
				return errors.WithStack(err)
			}
			deployment := s.registry.CreateDeployment(ctx, req.Key, module, existingModule, func(element *schema.RuntimeElement) error {
				cs := req.Key.String()
				_, err := s.schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{
					Changeset: &cs,
					Update:    element.ToProto(),
				}))
				if err != nil {
					return errors.Wrap(err, "error updating runtime")
				}
				return nil
			})
			if err := deployment.Run(ctx); err != nil {
				return errors.Wrap(err, "error running deployment")
			}
			logger.Debugf("Finished deployment for module %s", moduleName)
			return nil
		})

	}
	err := group.Wait()
	if err != nil {
		return errors.Wrap(err, "error running deployments")
	}

	changeset := req.Key.String()
	for _, mod := range req.InternalRealm().Modules {
		element := &schema.RuntimeElement{Deployment: mod.Runtime.Deployment.DeploymentKey, Element: &schema.ModuleRuntimeDeployment{DeploymentKey: mod.Runtime.Deployment.DeploymentKey, State: schema.DeploymentStateReady}}
		_, err = s.schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Changeset: &changeset, Update: element.ToProto()}))
		if err != nil {
			return errors.Wrap(err, "error preparing changeset")
		}
	}
	_, err = s.schemaClient.PrepareChangeset(ctx, connect.NewRequest(&ftlv1.PrepareChangesetRequest{Changeset: req.Key.String()}))
	if err != nil {
		return errors.Wrap(err, "error preparing changeset")
	}
	return nil
}

func (s *Service) handleScalingUpdate(ctx context.Context, dk key.Deployment, element *schema.ModuleRuntimeScaling) error {
	mLogger := log.FromContext(ctx)
	logger := mLogger.Module(dk.Payload.Module)
	logger.Debugf("Updating replicas %s to %d", dk.Payload.Module, element.MinReplicas)
	ctx = log.ContextWithLogger(ctx, logger)
	moduleName := dk.Payload.Module

	existingModule := s.eventSource.CanonicalView().Deployment(dk)
	if !existingModule.Ok() {
		return errors.Errorf("deployment %s not found", dk.String())
	}
	module := reflect.DeepCopy(existingModule.MustGet())
	module.Runtime.Scaling = element
	if err := s.registry.VerifyDeploymentSupported(ctx, module); err != nil {
		return errors.WithStack(err)
	}
	// TODO: fake changeset key
	deployment := s.registry.CreateDeployment(ctx, key.NewChangesetKey(), module, existingModule.MustGet(), func(element *schema.RuntimeElement) error {
		return nil
	})
	if err := deployment.Run(ctx); err != nil {
		return errors.Wrap(err, "error running deployment")
	}
	logger.Debugf("Finished deployment for module %s", moduleName)
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
