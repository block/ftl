package provisioner

import (
	"context"
	"fmt"
	"strings"

	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// provisionerPluginConfig is a map of provisioner name to resources it supports
type provisionerPluginConfig struct {
	// The default provisioner to use for all resources not matched here
	Default string `toml:"default"`
	Plugins []struct {
		ID        string                `toml:"id"`
		Resources []schema.ResourceType `toml:"resources"`
	} `toml:"plugins"`
}

func (cfg *provisionerPluginConfig) Validate() error {
	registeredResources := map[schema.ResourceType]bool{}
	for _, plugin := range cfg.Plugins {
		for _, r := range plugin.Resources {
			if registeredResources[r] {
				return fmt.Errorf("resource type %s is already registered. Trying to re-register for %s", r, plugin.ID)
			}
			registeredResources[r] = true
		}
	}
	return nil
}

// ProvisionerBinding is a Provisioner and the types it supports
type ProvisionerBinding struct {
	Provisioner provisionerconnect.ProvisionerPluginServiceClient
	ID          string
	Types       []schema.ResourceType
}

func (p ProvisionerBinding) String() string {
	types := []string{}
	for _, t := range p.Types {
		types = append(types, string(t))
	}
	return fmt.Sprintf("%s (%s)", p.ID, strings.Join(types, ","))
}

// ProvisionerRegistry contains all known resource handlers in the order they should be executed
type ProvisionerRegistry struct {
	Bindings []*ProvisionerBinding
}

// listBindings in the order they should be executed
func (reg *ProvisionerRegistry) listBindings() []*ProvisionerBinding {
	result := []*ProvisionerBinding{}
	result = append(result, reg.Bindings...)
	return result
}

func registryFromConfig(ctx context.Context, workingDir string, cfg *provisionerPluginConfig, runnerScaling scaling.RunnerScaling) (*ProvisionerRegistry, error) {
	logger := log.FromContext(ctx)
	result := &ProvisionerRegistry{}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating provisioner config: %w", err)
	}
	for _, plugin := range cfg.Plugins {
		provisioner, err := provisionerIDToProvisioner(ctx, plugin.ID, workingDir, runnerScaling)
		if err != nil {
			return nil, err
		}
		binding := result.Register(plugin.ID, provisioner, plugin.Resources...)
		logger.Debugf("Registered provisioner %s", binding)
	}
	return result, nil
}

func provisionerIDToProvisioner(ctx context.Context, id string, workingDir string, scaling scaling.RunnerScaling) (provisionerconnect.ProvisionerPluginServiceClient, error) {
	switch id {
	case "kubernetes":
		// TODO: move this into a plugin
		return NewRunnerScalingProvisioner(scaling), nil
	case "noop":
		return &NoopProvisioner{}, nil
	default:
		plugin, _, err := plugin.Spawn(
			ctx,
			log.FromContext(ctx).GetLevel(),
			"ftl-provisioner-"+id,
			"",
			workingDir,
			"ftl-provisioner-"+id,
			provisionerconnect.NewProvisionerPluginServiceClient,
		)
		if err != nil {
			return nil, fmt.Errorf("error spawning plugin: %w", err)
		}

		return plugin.Client, nil
	}
}

// Register to the registry, to be executed after all the previously added handlers
func (reg *ProvisionerRegistry) Register(id string, handler provisionerconnect.ProvisionerPluginServiceClient, types ...schema.ResourceType) *ProvisionerBinding {
	binding := &ProvisionerBinding{
		Provisioner: handler,
		Types:       types,
		ID:          id,
	}
	reg.Bindings = append(reg.Bindings, binding)
	return binding
}

// CreateDeployment to take the system to the desired state
func (reg *ProvisionerRegistry) CreateDeployment(ctx context.Context, changeset key.Changeset, desiredModule, existingModule *schema.Module, updateHandler func(*schema.RuntimeElement) error) *Deployment {
	logger := log.FromContext(ctx)
	module := desiredModule.GetName()

	deployment := &Deployment{
		DeploymentState: desiredModule,
		Previous:        existingModule,
		Changeset:       changeset,
		UpdateHandler:   updateHandler,
	}
	var allDesired, allExisting schema.ResourceSet
	allDesired = schema.GetProvisionedResources(desiredModule)
	allExisting = schema.GetProvisionedResources(existingModule)

	for _, binding := range reg.listBindings() {
		desired := allDesired.FilterByType(binding.Types...)
		existing := allExisting.FilterByType(binding.Types...)

		ds := false
		for _, r := range desired {
			if r.DeploymentSpecific {
				ds = true
				break
			}
		}

		for _, r := range existing {
			if r.DeploymentSpecific {
				ds = true
				break
			}
		}

		if !desired.IsEqual(existing) || ds {
			logger.Debugf("Adding task for module %s: %s", module, binding.ID)
			deployment.Tasks = append(deployment.Tasks, &Task{
				module:     module,
				binding:    binding,
				deployment: deployment,
			})
		}
	}
	return deployment
}
