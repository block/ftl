package main

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	provisionerpb "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/provisioner/state"
)

func (c *SandboxProvisioner) Status(ctx context.Context, req *connect.Request[provisionerpb.StatusRequest]) (*connect.Response[provisionerpb.StatusResponse], error) {
	token := req.Msg.ProvisioningToken
	// if the task is not in the map, it means that the provisioner has crashed since starting the task
	// in that case, we start a new task to query the existing stack
	task, loaded := c.running.Load(token)
	if !loaded {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("task %s not found", token))
	}

	if task.Err() != nil {
		c.running.Delete(token)
		return nil, connect.NewError(connect.CodeUnknown, task.Err())
	}

	outputs := task.Outputs()
	if outputs != nil {
		c.running.Delete(token)

		deploymentKey, err := key.ParseDeploymentKey(req.Msg.DesiredModule.Runtime.Deployment.DeploymentKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse deployment key: %w", err)
		}
		events, err := c.updateResources(deploymentKey, outputs)
		if err != nil {
			return nil, fmt.Errorf("failed to update resources: %w", err)
		}
		return connect.NewResponse(&provisionerpb.StatusResponse{
			Status: &provisionerpb.StatusResponse_Success{
				Success: &provisionerpb.StatusResponse_ProvisioningSuccess{
					Outputs: slices.Map(events, func(t *schema.RuntimeElement) *schemapb.RuntimeElement {
						return t.ToProto()
					}),
				},
			},
		}), nil
	}

	return connect.NewResponse(&provisionerpb.StatusResponse{
		Status: &provisionerpb.StatusResponse_Running{
			Running: &provisionerpb.StatusResponse_ProvisioningRunning{},
		},
	}), nil
}

func (c *SandboxProvisioner) updateResources(deployment key.Deployment, outputs []state.State) ([]*schema.RuntimeElement, error) {
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
		case state.OutputTopic:
			results = append(results, &schema.RuntimeElement{
				Deployment: deployment,
				Name:       optional.Some(o.Topic),
				Element:    o.Runtime,
			})
		case state.OutputSubscription:
			results = append(results, &schema.RuntimeElement{
				Deployment: deployment,
				Name:       optional.Some(o.Verb),
				Element:    o.Runtime,
			})
		default:
			return nil, fmt.Errorf("unknown output type: %T", o)
		}
	}

	return results, nil
}
