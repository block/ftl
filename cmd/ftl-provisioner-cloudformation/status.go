package main

import (
	"context"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"

	provisionerpb "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1"
	"github.com/block/ftl/common/key"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

func (c *CloudformationProvisioner) Status(ctx context.Context, req *connect.Request[provisionerpb.StatusRequest]) (*connect.Response[provisionerpb.StatusResponse], error) {
	token := req.Msg.ProvisioningToken
	// if the task is not in the map, it means that the provisioner has crashed since starting the task
	// in that case, we start a new task to query the existing stack
	task, loaded := c.running.LoadOrStore(token, &provisioner.Task{})
	if !loaded {
		dk, err := key.ParseDeploymentKey(token)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse deployment key")
		}
		task.Start(ctx, req.Msg.DesiredModule.Name, dk)
	}

	if task.Err() != nil {
		c.running.Delete(token)
		return nil, errors.WithStack(connect.NewError(connect.CodeUnknown, task.Err()))
	}

	outputs := task.Outputs()
	if outputs != nil {
		c.running.Delete(token)

		deploymentKey, err := key.ParseDeploymentKey(req.Msg.DesiredModule.Runtime.Deployment.DeploymentKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse deployment key")
		}
		events, err := c.updateResources(deploymentKey, outputs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to update resources")
		}
		return connect.NewResponse(&provisionerpb.StatusResponse{
			Status: &provisionerpb.ProvisioningStatus{
				Status: &provisionerpb.ProvisioningStatus_Success{
					Success: &provisionerpb.ProvisioningStatus_ProvisioningSuccess{
						Outputs: slices.Map(events, func(t *schema.RuntimeElement) *schemapb.RuntimeElement {
							return t.ToProto()
						}),
					}},
			},
		}), nil
	}

	return connect.NewResponse(&provisionerpb.StatusResponse{
		Status: &provisionerpb.ProvisioningStatus{
			Status: &provisionerpb.ProvisioningStatus_Running{
				Running: &provisionerpb.ProvisioningStatus_ProvisioningRunning{
					ProvisioningToken: token,
				},
			},
		},
	}), nil
}

func outputsByResourceID(outputs []types.Output) (map[string][]types.Output, error) {
	m := make(map[string][]types.Output)
	for _, output := range outputs {
		key, err := decodeOutputKey(output)
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode output key")
		}
		m[key.ResourceID] = append(m[key.ResourceID], output)
	}
	return m, nil
}

func (c *CloudformationProvisioner) updateResources(deployment key.Deployment, outputs []state.State) ([]*schema.RuntimeElement, error) {
	var results []*schema.RuntimeElement

	for _, output := range outputs {
		switch o := output.(type) {
		case state.OutputPostgres:
			results = append(results, &schema.RuntimeElement{
				Deployment: deployment,
				Name:       optional.Some(o.ResourceID),
				Element: &schema.DatabaseRuntime{
					Connections: &schema.DatabaseRuntimeConnections{
						Write: o.Connector,
						Read:  o.Connector,
					},
				},
			})
		case state.OutputMySQL:
			results = append(results, &schema.RuntimeElement{
				Deployment: deployment,
				Name:       optional.Some(o.ResourceID),
				Element: &schema.DatabaseRuntime{
					Connections: &schema.DatabaseRuntimeConnections{
						Write: o.Connector,
						Read:  o.Connector,
					},
				},
			})
		default:
			return nil, errors.Errorf("unknown output type: %T", o)
		}
	}

	return results, nil
}
