package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	client, err := createClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudformation client: %w", err)
	}

	desc, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: &req.Msg.ProvisioningToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe stack: %w", err)
	}
	stack := desc.Stacks[0]

	switch stack.StackStatus {
	case types.StackStatusCreateInProgress:
		return running()
	case types.StackStatusCreateFailed:
		return failure(&stack)
	case types.StackStatusCreateComplete:
		return c.success(ctx, &stack, req.Msg.DesiredResources)
	case types.StackStatusRollbackInProgress:
		return failure(&stack)
	case types.StackStatusRollbackFailed:
		return failure(&stack)
	case types.StackStatusRollbackComplete:
		return failure(&stack)
	case types.StackStatusDeleteInProgress:
		return running()
	case types.StackStatusDeleteFailed:
		return failure(&stack)
	case types.StackStatusDeleteComplete:
		return c.success(ctx, &stack, req.Msg.DesiredResources)
	case types.StackStatusUpdateInProgress:
		return running()
	case types.StackStatusUpdateCompleteCleanupInProgress:
		return running()
	case types.StackStatusUpdateComplete:
		return c.success(ctx, &stack, req.Msg.DesiredResources)
	case types.StackStatusUpdateFailed:
		return failure(&stack)
	case types.StackStatusUpdateRollbackInProgress:
		return running()
	default:
		return nil, errors.New("unsupported Cloudformation status code: " + string(desc.Stacks[0].StackStatus))
	}
}

func (c *CloudformationProvisioner) success(ctx context.Context, stack *types.Stack, resources []*provisioner.Resource) (*connect.Response[provisioner.StatusResponse], error) {
	err := c.updateResources(ctx, stack.Outputs, resources)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Success{
			Success: &provisioner.StatusResponse_ProvisioningSuccess{
				UpdatedResources: resources,
			},
		},
	}), nil
}

func running() (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.StatusResponse_Running{
			Running: &provisioner.StatusResponse_ProvisioningRunning{},
		},
	}), nil
}

func failure(stack *types.Stack) (*connect.Response[provisioner.StatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnknown, errors.New(*stack.StackStatusReason))
}

func outputsByResourceID(outputs []types.Output) (map[string][]types.Output, error) {
	m := make(map[string][]types.Output)
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return nil, fmt.Errorf("failed to decode output key: %w", err)
		}
		m[key.ResourceID] = append(m[key.ResourceID], output)
	}
	return m, nil
}

func outputsByPropertyName(outputs []types.Output) (map[string]types.Output, error) {
	m := make(map[string]types.Output)
	for _, output := range outputs {
		m[*output.OutputKey] = output
	}
	return m, nil
}

func (c *CloudformationProvisioner) updateResources(ctx context.Context, outputs []types.Output, update []*provisioner.Resource) error {
	byResourceID, err := outputsByResourceID(outputs)
	if err != nil {
		return fmt.Errorf("failed to group outputs by resource ID: %w", err)
	}

	for _, resource := range update {
		if postgres, ok := resource.Resource.(*provisioner.Resource_Postgres); ok {
			if postgres.Postgres == nil {
				postgres.Postgres = &provisioner.PostgresResource{}
			}
			if postgres.Postgres.Output == nil {
				postgres.Postgres.Output = &provisioner.PostgresResource_PostgresResourceOutput{}
			}

			if err := c.updatePostgresOutputs(ctx, postgres.Postgres.Output, resource.ResourceId, byResourceID[resource.ResourceId]); err != nil {
				return fmt.Errorf("failed to update postgres outputs: %w", err)
			}
		} else if _, ok := resource.Resource.(*provisioner.Resource_Mysql); ok {
			panic("mysql not implemented")
		}
	}
	return nil
}

func (c *CloudformationProvisioner) updatePostgresOutputs(ctx context.Context, to *provisioner.PostgresResource_PostgresResourceOutput, resourceID string, outputs []types.Output) error {
	byName, err := outputsByPropertyName(outputs)
	if err != nil {
		return fmt.Errorf("failed to group outputs by property name: %w", err)
	}

	secretARN := *byName[PropertyMasterUserSecretARN].OutputValue
	username, password, err := c.secretARNToUsernamePassword(ctx, secretARN)
	if err != nil {
		return fmt.Errorf("failed to get username and password from secret ARN: %w", err)
	}

	to.ReadDsn = endpointToDSN(*byName[PropertyDBReadEndpoint].OutputValue, resourceID, 5432, username, password)
	to.WriteDsn = endpointToDSN(*byName[PropertyDBWriteEndpoint].OutputValue, resourceID, 5432, username, password)

	return nil
}

func endpointToDSN(endpoint, database string, port int, username, password string) string {
	urlEncodedPassword := url.QueryEscape(password)
	return fmt.Sprintf("postgres://%s:%d/%s?user=%s&password=%s", endpoint, port, database, username, urlEncodedPassword)
}

func (c *CloudformationProvisioner) secretARNToUsernamePassword(ctx context.Context, secretARN string) (string, string, error) {
	secret, err := c.secrets.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret value: %w", err)
	}
	secretString := *secret.SecretString

	var secretData map[string]string
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	return secretData["username"], secretData["password"], nil
}
