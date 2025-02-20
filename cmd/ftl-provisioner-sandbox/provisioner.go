package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/google/uuid"
	"github.com/puzpuzpuz/xsync/v3"

	provisionerpb "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/executor"
	"github.com/block/ftl/internal/provisioner/state"
)

type Config struct {
	MySQLCredentialsSecretARN string `help:"ARN for the secret containing mysql credentials" env:"FTL_SANDBOX_MYSQL_ARN"`
	MySQLEndpoint             string `help:"Endpoint for the mysql database" env:"FTL_SANDBOX_MYSQL_ENDPOINT"`
}

type SandboxProvisioner struct {
	secrets *secretsmanager.Client
	confg   *Config

	running *xsync.MapOf[string, *task]
}

var _ provisionerconnect.ProvisionerPluginServiceHandler = (*SandboxProvisioner)(nil)

func NewSandboxProvisioner(ctx context.Context, config Config) (context.Context, *SandboxProvisioner, error) {
	secrets, err := createSecretsClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create secretsmanager client: %w", err)
	}

	return ctx, &SandboxProvisioner{
		secrets: secrets,
		confg:   &config,
		running: xsync.NewMapOf[string, *task](),
	}, nil
}

func (c *SandboxProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (c *SandboxProvisioner) Provision(ctx context.Context, req *connect.Request[provisionerpb.ProvisionRequest]) (*connect.Response[provisionerpb.ProvisionResponse], error) {
	logger := log.FromContext(ctx)

	module, err := schema.ModuleFromProto(req.Msg.DesiredModule)
	if err != nil {
		return nil, fmt.Errorf("failed to convert module from proto: %w", err)
	}
	var acceptedKinds []schema.ResourceType
	for _, k := range req.Msg.Kinds {
		acceptedKinds = append(acceptedKinds, schema.ResourceType(k))
	}

	inputStates := inputsFromSchema(module, acceptedKinds, req.Msg.DesiredModule.Name, c.confg)

	token := uuid.New().String()

	runner := &provisioner.Runner{
		State: inputStates,
		Stages: []provisioner.RunnerStage{
			{
				Name: "infrastructure-setup",
				Handlers: []provisioner.Handler{{
					Executor: executor.NewARNSecretMySQLSetup(c.secrets, req.Msg.DesiredModule.Name),
					Handles:  []state.State{state.RDSInstanceReadyMySQL{}},
				}},
			},
		},
	}

	task := &task{runner: runner}
	if _, ok := c.running.LoadOrStore(token, task); ok {
		return nil, fmt.Errorf("provisioner already running: %s", token)
	}
	logger.Debugf("Starting task for module %s: %s", req.Msg.DesiredModule.Name, token)
	task.Start(ctx)
	return connect.NewResponse(&provisionerpb.ProvisionResponse{
		Status:            provisionerpb.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
		ProvisioningToken: token,
	}), nil
}

func createSecretsClient(ctx context.Context) (*secretsmanager.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}
	return secretsmanager.New(
		secretsmanager.Options{
			Credentials: cfg.Credentials,
			Region:      cfg.Region,
		},
	), nil
}

func inputsFromSchema(module *schema.Module, acceptedKinds []schema.ResourceType, moduleName string, config *Config) []state.State {
	var inputStates []state.State
	for _, provisioned := range schema.GetProvisioned(module) {
		for _, resource := range provisioned.GetProvisioned().FilterByType(acceptedKinds...) {
			switch resource.Kind {
			case schema.ResourceTypeMysql:
				input := state.RDSInstanceReadyMySQL{
					ResourceID:          provisioned.ResourceID(),
					Module:              moduleName,
					MasterUserSecretARN: config.MySQLCredentialsSecretARN,
					WriteEndpoint:       config.MySQLEndpoint,
					ReadEndpoint:        config.MySQLEndpoint,
				}
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
