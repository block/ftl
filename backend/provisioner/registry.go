package provisioner

import (
	"context"
	"fmt"
	"strings"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1/provisionerpbconnect"
	"github.com/block/ftl/backend/provisioner/scaling"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/plugin"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/oci"
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
				return errors.Errorf("resource type %s is already registered. Trying to re-register for %s", r, plugin.ID)
			}
			registeredResources[r] = true
		}
	}
	return nil
}

// ProvisionerBinding is a Provisioner and the types it supports
type ProvisionerBinding struct {
	Provisioner Plugin
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

func registryFromConfig(ctx context.Context, workingDir string, cfg *provisionerPluginConfig, runnerScaling scaling.RunnerScaling, adminClient adminpbconnect.AdminServiceClient, imageService *oci.ImageService) (*ProvisionerRegistry, error) {
	logger := log.FromContext(ctx)
	result := &ProvisionerRegistry{}
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "error validating provisioner config")
	}
	for _, plugin := range cfg.Plugins {
		provisioner, err := provisionerIDToProvisioner(ctx, plugin.ID, workingDir, runnerScaling, adminClient, imageService)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		binding := result.Register(plugin.ID, provisioner, plugin.Resources...)
		logger.Debugf("Registered provisioner %s", binding)
	}
	return result, nil
}

func provisionerIDToProvisioner(ctx context.Context, id string, workingDir string, scaling scaling.RunnerScaling, adminClient adminpbconnect.AdminServiceClient, imageService *oci.ImageService) (Plugin, error) {
	switch id {
	case "kubernetes":
		// TODO: move this into a plugin
		return NewRunnerScalingProvisioner(scaling, false), nil
	case "simple-egress":
		// TODO: move this into a plugin
		return NewEgressProvisioner(adminClient), nil
	case "noop":
		return NewPluginClient(&NoopProvisioner{}), nil
	case "oci-image":
		return NewOCIImageProvisioner(imageService, "ftl0/ftl-runner"), nil
	default:
		plugin, _, err := plugin.Spawn(
			ctx,
			log.FromContext(ctx).GetLevel(),
			"ftl-provisioner-"+id,
			"",
			workingDir,
			"ftl-provisioner-"+id,
			provisionerconnect.NewProvisionerPluginServiceClient,
			true,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error spawning plugin")
		}

		return NewPluginClient(plugin.Client), nil
	}
}

// Register to the registry, to be executed after all the previously added handlers
func (reg *ProvisionerRegistry) Register(id string, handler Plugin, types ...schema.ResourceType) *ProvisionerBinding {
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

func (reg *ProvisionerRegistry) VerifyDeploymentSupported(ctx context.Context, module *schema.Module) error {
	for _, r := range schema.GetProvisionedResources(module) {
		supported := false
		for _, binding := range reg.Bindings {
			if slices.Contains(binding.Types, r.Kind) {
				supported = true
				break
			}
		}
		if !supported {
			return errors.Errorf("resource type %s is not supported in this environment", r.Kind)
		}
	}
	return nil
}
