package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/puzpuzpuz/xsync/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

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

const (
	PropertyPsqlReadEndpoint   = "psql:read_endpoint"
	PropertyPsqlWriteEndpoint  = "psql:write_endpoint"
	PropertyPsqlMasterUserARN  = "psql:master_user_secret_arn"
	PropertyMySQLReadEndpoint  = "mysql:read_endpoint"
	PropertyMySQLWriteEndpoint = "mysql:write_endpoint"
	PropertyMySQLMasterUserARN = "mysql:master_user_secret_arn"
)

type Config struct {
	DatabaseSubnetGroupARN string `help:"ARN for the subnet group to be used to create Databases in" env:"FTL_PROVISIONER_CF_DB_SUBNET_GROUP"`
	// TODO: remove this once we have module specific security groups
	DatabaseSecurityGroup string `help:"SG for databases" env:"FTL_PROVISIONER_CF_DB_SECURITY_GROUP"`
	MysqlSecurityGroup    string `help:"SG for mysql" env:"FTL_PROVISIONER_CF_MYSQL_SECURITY_GROUP"`
}

type CloudformationProvisioner struct {
	client  *cloudformation.Client
	secrets *secretsmanager.Client
	confg   *Config

	running *xsync.MapOf[string, *provisioner.Task]
}

var _ provisionerconnect.ProvisionerPluginServiceHandler = (*CloudformationProvisioner)(nil)

func NewCloudformationProvisioner(ctx context.Context, config Config) (context.Context, *CloudformationProvisioner, error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cloudformation client: %w", err)
	}
	secrets, err := createSecretsClient(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create secretsmanager client: %w", err)
	}

	return ctx, &CloudformationProvisioner{
		client:  client,
		secrets: secrets,
		confg:   &config,
		running: xsync.NewMapOf[string, *provisioner.Task](),
	}, nil
}

func (c *CloudformationProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (c *CloudformationProvisioner) Provision(ctx context.Context, req *connect.Request[provisionerpb.ProvisionRequest]) (*connect.Response[provisionerpb.ProvisionResponse], error) {
	logger := log.FromContext(ctx)

	module, err := schema.ModuleFromProto(req.Msg.DesiredModule)
	if err != nil {
		return nil, fmt.Errorf("failed to convert module from proto: %w", err)
	}
	var acceptedKinds []schema.ResourceType
	for _, k := range req.Msg.Kinds {
		acceptedKinds = append(acceptedKinds, schema.ResourceType(k))
	}

	inputStates := inputsFromSchema(module, acceptedKinds, req.Msg.FtlClusterId, req.Msg.DesiredModule.Name)
	stackID := stackName(req.Msg)

	runner := &provisioner.Runner{
		State: inputStates,
		Stages: []provisioner.RunnerStage{
			{
				Name: "cloudformation-update",
				Handlers: []provisioner.Handler{{
					Executor: NewCloudFormationExecutor(stackID, c.client, c.secrets, c.confg),
					Handles:  []state.State{state.InputPostgres{}, state.InputMySQL{}},
				}},
			}, {
				Name: "infrastructure-setup",
				Handlers: []provisioner.Handler{{
					Executor: executor.NewPostgresSetup(c.secrets),
					Handles:  []state.State{state.RDSInstanceReadyPostgres{}},
				}, {
					Executor: executor.NewARNSecretMySQLSetup(c.secrets, req.Msg.DesiredModule.Name),
					Handles:  []state.State{state.RDSInstanceReadyMySQL{}},
				}},
			},
		},
	}

	task := runner.AsyncTask()
	if _, ok := c.running.LoadOrStore(stackID, task); ok {
		return nil, fmt.Errorf("provisioner already running: %s", stackID)
	}
	logger.Debugf("Starting task for module %s: %s", req.Msg.DesiredModule.Name, stackID)
	task.Start(ctx, module.Name)
	return connect.NewResponse(&provisionerpb.ProvisionResponse{
		Status:            provisionerpb.ProvisionResponse_PROVISION_RESPONSE_STATUS_SUBMITTED,
		ProvisioningToken: stackID,
	}), nil
}

func stackName(req *provisionerpb.ProvisionRequest) string {
	return sanitize(req.FtlClusterId) + "-" + sanitize(req.DesiredModule.Name)
}

func generateChangeSetName(stack string) string {
	return sanitize(stack) + strconv.FormatInt(time.Now().Unix(), 10)
}

func inputsFromSchema(module *schema.Module, acceptedKinds []schema.ResourceType, clusterID string, moduleName string) []state.State {
	var inputStates []state.State
	for _, provisioned := range schema.GetProvisioned(module) {
		for _, resource := range provisioned.GetProvisioned().FilterByType(acceptedKinds...) {
			switch resource.Kind {
			case schema.ResourceTypePostgres:
				input := state.InputPostgres{
					ResourceID: provisioned.ResourceID(),
					Cluster:    clusterID,
					Module:     moduleName,
				}
				inputStates = append(inputStates, input)
			case schema.ResourceTypeMysql:
				input := state.InputMySQL{
					ResourceID: provisioned.ResourceID(),
					Cluster:    clusterID,
					Module:     moduleName,
				}
				inputStates = append(inputStates, input)
			default:
				continue
			}
		}
	}
	return inputStates
}

// ResourceTemplater interface for different resource types
type ResourceTemplater interface {
	AddToTemplate(tmpl *goformation.Template) error
}

func ftlTags(cluster, module string) []tags.Tag {
	return []tags.Tag{{
		Key:   "ftl:module",
		Value: module,
	}, {
		Key:   "ftl:cluster",
		Value: cluster,
	}}
}

func cloudformationResourceID(strs ...string) string {
	caser := cases.Title(language.English)
	var buffer bytes.Buffer

	for _, s := range strs {
		buffer.WriteString(caser.String(s))
	}
	return buffer.String()
}

func sanitize(name string) string {
	// just keep alpha numeric chars
	s := []byte(name)
	j := 0
	for _, b := range s {
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			s[j] = b
			j++
		}
	}
	return string(s[:j])
}

func main() {
	plugin.Start(
		context.Background(),
		"ftl-provisioner-cloudformation",
		NewCloudformationProvisioner,
		"",
		provisionerconnect.NewProvisionerPluginServiceHandler,
	)
}

func ptr[T any](s T) *T { return &s }
