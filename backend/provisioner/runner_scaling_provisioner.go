package provisioner

import (
	"context"
	"fmt"
	"math"

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
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, _ *schema.Module) (*schema.RuntimeElement, error) {
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
		runner := false
		subscriptionCount := 0
		for _, decl := range module.Decls {
			if verb, ok := decl.(*schema.Verb); ok {
				// TODO: should cron be part of the subscription runner
				if verb.Export || verb.GetMetadataIngress().Ok() || verb.GetMetadataCronJob().Ok() {
					runner = true
				}
				if sub, ok := verb.GetMetadataSubscriber().Get(); ok {
					// TODO: we need to get the partition count here
					subscriptionCount = 1
				}
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
		element := &schema.ModuleRuntimeRunner{}
		if runner {
			endpointURI, err := scaling.StartDeployment(ctx, deployment.String(), module, cron, http, false)
			if err != nil {
				return nil, fmt.Errorf("failed to start deployment: %w", err)
			}
			element.Endpoint = endpointURI.String()
		} else {
			element.RunnerNotRequired = true
		}

		return &element, nil
	}
}

func deProvisionRunner(scaling scaling.RunnerScaling) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, rc schema.Provisioned, _ *schema.Module) (*schema.RuntimeElement, error) {
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
		logger.Debugf("Removing runner: %s for deployment %s", module.Name, deployment)
		err := scaling.TerminateDeployment(ctx, deployment.String())
		if err != nil {
			return nil, fmt.Errorf("failed to start deployment: %w", err)
		}
		return nil, nil
	}
}
