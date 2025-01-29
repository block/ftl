package provisioner

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	_ "github.com/go-sql-driver/mysql"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/provisioner/scaling"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

// NewRunnerScalingProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode

func NewRunnerScalingProvisioner(runners scaling.RunnerScaling) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeRunner: provisionRunner(runners),
	})
}

func provisionRunner(scaling scaling.RunnerScaling) InMemResourceProvisionerFn {
	return func(ctx context.Context, moduleName string, rc schema.Provisioned) (schema.Event, error) {
		logger := log.FromContext(ctx)

		module, ok := rc.(*schema.Module)
		if !ok {
			return nil, fmt.Errorf("expected module, got %T", rc)
		}

		deployment := module.Runtime.Deployment.DeploymentKey
		if deployment.IsZero() {
			return nil, fmt.Errorf("failed to find deployment for runner")
		}
		logger.Debugf("provisioning runner: %s.%s for deployment %s", module, rc.ResourceID(), deployment)
		cron := false
		http := false
		for _, decl := range module.Decls {
			if verb, ok := decl.(*schema.Verb); ok {
				for _, meta := range verb.Metadata {
					switch meta.(type) {
					case *schema.MetadataCronJob:
						cron = true
					case *schema.MetadataIngress:
						http = true
					default:

					}

				}
			}
		}
		if err := scaling.StartDeployment(ctx, module.Name, deployment.String(), module, cron, http); err != nil {
			logger.Infof("failed to start deployment: %v", err)
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}
		endpoint, err := scaling.GetEndpointForDeployment(ctx, module.Name, deployment.String())
		if err != nil || !endpoint.Ok() {
			return nil, fmt.Errorf("failed to get endpoint for deployment: %w", err)
		}
		ep := endpoint.MustGet()
		endpointURI := ep.String()

		runnerClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, endpointURI, log.Error)
		// TODO: a proper timeout
		timeout := time.After(1 * time.Minute)
		for {
			_, err := runnerClient.Ping(ctx, connect.NewRequest(&ftlv1.PingRequest{}))
			if err == nil {
				break
			}
			logger.Tracef("waiting for runner to be ready: %v", err)
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled %w", ctx.Err())
			case <-timeout:
				return nil, fmt.Errorf("timed out waiting for runner to be ready")
			case <-time.After(time.Millisecond * 100):
			}
		}

		schemaClient := rpc.ClientFromContext[ftlv1connect.SchemaServiceClient](ctx)

		deps, err := scaling.TerminatePreviousDeployments(ctx, module.Name, deployment.String())
		if err != nil {
			logger.Errorf(err, "failed to terminate previous deployments")
		} else {
			for _, dep := range deps {
				_, err = schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Deployment: deployment.String(), Event: &schemapb.ModuleRuntimeEvent{Scaling: &schemapb.ModuleRuntimeScaling{MinReplicas: 0}}}))
				if err != nil {
					logger.Errorf(err, "failed to update deployment %s", dep)
				}
			}
		}

		logger.Debugf("updating module runtime for %s with endpoint %s", module, endpointURI)
		dk := deployment.String()
		_, err = schemaClient.UpdateDeploymentRuntime(ctx, connect.NewRequest(&ftlv1.UpdateDeploymentRuntimeRequest{Event: &schemapb.ModuleRuntimeEvent{
			DeploymentKey: dk,
			Deployment: &schemapb.ModuleRuntimeDeployment{
				DeploymentKey: deployment.String(),
				Endpoint:      endpointURI,
			},
		}}))
		if err != nil {
			return nil, fmt.Errorf("failed to update module runtime: %w", err)
		}
		return &schema.ModuleRuntimeEvent{
			DeploymentKey: deployment,
			Deployment: optional.Some(schema.ModuleRuntimeDeployment{
				DeploymentKey: deployment,
				Endpoint:      endpointURI,
			}),
		}, nil
	}
}
