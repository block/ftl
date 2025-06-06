package main

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"

	provisionerpb "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1/provisionerpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/executor"
	"github.com/block/ftl/internal/provisioner/state"
	timeline "github.com/block/ftl/internal/timelineclient"
)

type Config struct {
	MySQLCredentialsSecretARN string   `help:"ARN for the secret containing mysql credentials" env:"FTL_SANDBOX_MYSQL_ARN"`
	MySQLEndpoint             string   `help:"Endpoint for the mysql database" env:"FTL_SANDBOX_MYSQL_ENDPOINT"`
	KafkaBrokers              []string `help:"Brokers for the kafka cluster" env:"FTL_SANDBOX_KAFKA_BROKERS"`

	TimelineConfig timeline.Config `embed:""`
}

type SandboxProvisioner struct {
	secrets *secretsmanager.Client
	confg   *Config

	running *xsync.MapOf[string, *provisioner.Task]
}

var _ provisionerconnect.ProvisionerPluginServiceHandler = (*SandboxProvisioner)(nil)

func NewSandboxProvisioner(ctx context.Context, config Config) (context.Context, *SandboxProvisioner, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Creating sandbox provisioner")

	secrets, err := createSecretsClient(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create secretsmanager client")
	}

	timelineClient := timeline.NewClient(ctx, config.TimelineConfig)
	timelineLogSink := timeline.NewLogSink(timelineClient, log.Debug)
	go timelineLogSink.RunLogLoop(ctx)
	logger = logger.AddSink(timelineLogSink)
	ctx = log.ContextWithLogger(ctx, logger)

	return ctx, &SandboxProvisioner{
		secrets: secrets,
		confg:   &config,
		running: xsync.NewMapOf[string, *provisioner.Task](),
	}, nil
}

func (c *SandboxProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (c *SandboxProvisioner) Provision(ctx context.Context, req *connect.Request[provisionerpb.ProvisionRequest]) (*connect.Response[provisionerpb.ProvisionResponse], error) {
	module, err := schema.ModuleFromProto(req.Msg.DesiredModule)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert module from proto")
	}

	logger := log.FromContext(ctx).Deployment(module.Runtime.Deployment.DeploymentKey).Module(module.Name)
	ctx = log.ContextWithLogger(ctx, logger)

	var acceptedKinds []schema.ResourceType
	for _, k := range req.Msg.Kinds {
		acceptedKinds = append(acceptedKinds, schema.ResourceType(k))
	}

	inputStates := inputsFromSchema(ctx, module, acceptedKinds, req.Msg.DesiredModule.Name, c.confg)

	token := uuid.New().String()

	runner := &provisioner.Runner{
		State: inputStates,
		Stages: []provisioner.RunnerStage{{
			Name: "infrastructure-setup",
			Handlers: []provisioner.Handler{{
				Executor: executor.NewARNSecretMySQLSetup(c.secrets, req.Msg.DesiredModule.Name),
				Handles:  []state.State{state.RDSInstanceReadyMySQL{}},
			}, {
				Executor: executor.NewKafkaTopicSetup(),
				Handles:  []state.State{state.TopicClusterReady{}},
			}},
		}},
	}

	task := runner.AsyncTask()
	if _, ok := c.running.LoadOrStore(token, task); ok {
		return nil, errors.Errorf("provisioner already running: %s", token)
	}
	logger.Debugf("Starting task %s", token)
	task.Start(ctx, module.Name, module.Runtime.Deployment.DeploymentKey)
	return connect.NewResponse(&provisionerpb.ProvisionResponse{
		Tasks: []*provisionerpb.ProvisioningStatus{{
			Status: &provisionerpb.ProvisioningStatus_Running{
				Running: &provisionerpb.ProvisioningStatus_ProvisioningRunning{
					ProvisioningToken: token,
				},
			},
		}},
	}), nil
}

func createSecretsClient(ctx context.Context) (*secretsmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load default aws config")
	}
	return secretsmanager.New(
		secretsmanager.Options{
			Credentials: cfg.Credentials,
			Region:      cfg.Region,
		},
	), nil
}

func inputsFromSchema(
	ctx context.Context,
	module *schema.Module,
	acceptedKinds []schema.ResourceType,
	moduleName string,
	config *Config,
) []state.State {
	logger := log.FromContext(ctx).Module(module.Name)
	logger.Debugf("Reading inputs from schema for kinds: %v", acceptedKinds)

	var inputStates []state.State
	for _, provisioned := range schema.GetProvisioned(module) {
		for _, resource := range provisioned.GetProvisioned().FilterByType(acceptedKinds...) {
			switch resource.Kind {
			case schema.ResourceTypeMysql:
				input := state.RDSInstanceReadyMySQL{
					ResourceID:          provisioned.ResourceID(),
					Module:              moduleName,
					MasterUserSecretARN: optional.From(config.MySQLCredentialsSecretARN, config.MySQLCredentialsSecretARN != ""),
					WriteEndpoint:       config.MySQLEndpoint,
					ReadEndpoint:        config.MySQLEndpoint,
				}
				logger.Debugf("Adding %s", input.DebugString())
				inputStates = append(inputStates, input)
			case schema.ResourceTypeTopic:
				topic, ok := (provisioned).(*schema.Topic)
				if !ok {
					logger.Warnf("Skipping non-topic resource %s", provisioned.ResourceID())
					continue
				}
				partitions := 1
				if pm, ok := slices.FindVariant[*schema.MetadataPartitions](topic.Metadata); ok {
					partitions = pm.Partitions
				}
				input := state.TopicClusterReady{
					InputTopic: state.InputTopic{
						Topic:      topic.Name,
						Module:     moduleName,
						Partitions: partitions,
					},
					Brokers: config.KafkaBrokers,
				}
				logger.Debugf("Adding %s", input.DebugString())
				inputStates = append(inputStates, input)
			case schema.ResourceTypeSubscription:
				verb, ok := (provisioned).(*schema.Verb)
				if !ok {
					logger.Warnf("Skipping non-verb resource %s", provisioned.ResourceID())
					continue
				}
				// There is no work needed for subscriptions, so we just place the output state here
				input := state.OutputSubscription{
					Module: moduleName,
					Verb:   verb.Name,
					Connector: &schema.PlaintextKafkaSubscriptionConnector{
						KafkaBrokers: config.KafkaBrokers,
					},
				}
				logger.Debugf("Adding %s", input.DebugString())
				inputStates = append(inputStates, input)
			default:
				continue
			}
		}
	}
	return inputStates
}

func main() {
	plugin.Start(
		context.Background(),
		"ftl-provisioner-sandbox",
		NewSandboxProvisioner,
		"",
		provisionerconnect.NewProvisionerPluginServiceHandler,
	)
}
