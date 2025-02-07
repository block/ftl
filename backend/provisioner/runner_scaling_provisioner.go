package provisioner

import (
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// NewRunnerScalingProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode

func NewRunnerScalingProvisioner(runners scaling.RunnerScaling) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeRunner: provisionRunner(runners),
	},
		map[schema.ResourceType]InMemResourceProvisionerFn{
			schema.ResourceTypeRunner: deProvisionRunner(runners),
		})
}

func provisionRunner(scaling scaling.RunnerScaling) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned) (*schema.RuntimeElement, error) {
		if changeset.IsZero() {
			return nil, fmt.Errorf("changeset must be provided")
		}
		logger := log.FromContext(ctx)

		module, ok := rc.(*schema.Module)
		if !ok {
			return nil, fmt.Errorf("expected module, got %T", rc)
		}

		if deployment.IsZero() {
			return nil, fmt.Errorf("failed to find deployment for runner")
		}
		logger.Debugf("Provisioning runner: %s for deployment %s", module.Name, deployment)
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
		endpointURI, err := scaling.StartDeployment(ctx, deployment.String(), module, cron, http)
		if err != nil {
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}

		return &schema.RuntimeElement{
			Deployment: deployment,
			Element: &schema.ModuleRuntimeRunner{
				Endpoint: endpointURI.String(),
			},
		}, nil
	}
}

func deProvisionRunner(scaling scaling.RunnerScaling) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned) (*schema.RuntimeElement, error) {
		if changeset.IsZero() {
			return nil, fmt.Errorf("changeset must be provided")
		}
		logger := log.FromContext(ctx)

		module, ok := rc.(*schema.Module)
		if !ok {
			return nil, fmt.Errorf("expected module, got %T", rc)
		}

		if deployment.IsZero() {
			return nil, fmt.Errorf("failed to find deployment for runner")
		}
		logger.Infof("Removing runner: %s for deployment %s", module.Name, deployment)
		err := scaling.TerminateDeployment(ctx, deployment.String())
		if err != nil {
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}
		return nil, err
	}
}
