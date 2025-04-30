package provisioner

import (
	"context"

	errors "github.com/alecthomas/errors"
	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// NewRunnerScalingProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode

func NewRunnerScalingProvisioner(runners scaling.RunnerScaling, local bool) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypeRunner: provisionRunner(runners, local),
	},
		map[schema.ResourceType]InMemResourceProvisionerFn{
			schema.ResourceTypeRunner: deProvisionRunner(runners),
		})
}

func provisionRunner(scaling scaling.RunnerScaling, local bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, _ *schema.Module) (*schema.RuntimeElement, error) {
		if changeset.IsZero() {
			return nil, errors.Errorf("changeset must be provided")
		}
		logger := log.FromContext(ctx)

		module, ok := rc.(*schema.Module)
		if !ok {
			return nil, errors.Errorf("expected module, got %T", rc)
		}

		if deployment.IsZero() {
			return nil, errors.Errorf("failed to find deployment for runner")
		}
		logger.Debugf("Provisioning runner: %s for deployment %s", module.Name, deployment)
		cron := false
		http := false
		if !local {
			// We don't create kube deployments if there are no verbs
			// We still do for local dev as this can mess with JVM hot reload if there is no runner to talk to
			runner := false
			for _, decl := range module.Decls {
				if verb, ok := decl.(*schema.Verb); ok {
					runner = true
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
			if !runner {
				return &schema.RuntimeElement{
					Deployment: deployment,
					Element: &schema.ModuleRuntimeRunner{
						RunnerNotRequired: true,
					},
				}, nil
			}
		}
		endpointURI, err := scaling.StartDeployment(ctx, deployment.String(), module, cron, http)
		if err != nil {
			return nil, errors.Wrap(err, "failed to start deployment")
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
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, _ *schema.Module) (*schema.RuntimeElement, error) {
		if changeset.IsZero() {
			return nil, errors.Errorf("changeset must be provided")
		}
		logger := log.FromContext(ctx)

		module, ok := rc.(*schema.Module)
		if !ok {
			return nil, errors.Errorf("expected module, got %T", rc)
		}

		if deployment.IsZero() {
			return nil, errors.Errorf("failed to find deployment for runner")
		}
		logger.Debugf("Removing runner: %s for deployment %s", module.Name, deployment)
		err := scaling.TerminateDeployment(ctx, deployment.String())
		if err != nil {
			return nil, errors.Wrap(err, "failed to start deployment")
		}
		return nil, nil
	}
}
