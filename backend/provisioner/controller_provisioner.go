package provisioner

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/alecthomas/types/optional"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// NewControllerProvisioner creates a new provisioner that uses the FTL controller to provision modules
func NewControllerProvisioner(client ftlv1connect.ControllerServiceClient) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeModule: func(ctx context.Context, moduleName string, res schema.Provisioned) (schema.Event, error) {
			logger := log.FromContext(ctx)
			logger.Debugf("Provisioning module: %s", moduleName)

			module, ok := res.(*schema.Module)
			if !ok {
				return nil, fmt.Errorf("expected module, got %T", res)
			}

			resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
				Schema: module.ToProto(),
			}))
			if err != nil {
				return nil, fmt.Errorf("failed to create deployment: %w", err)
			}
			deploymentKey, err := key.ParseDeploymentKey(resp.Msg.DeploymentKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse deployment key: %w", err)
			}

			return &schema.ModuleRuntimeEvent{
				Module: module.Name,
				Deployment: optional.Some(schema.ModuleRuntimeDeployment{
					DeploymentKey: deploymentKey,
				}),
			}, nil
		},
	})
}
