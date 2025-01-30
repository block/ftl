package schemaservice

import (
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
)

// TODO: these should be event methods once we can move them to this package

// ApplyEvent applies an event to the schema state
func (r SchemaState) ApplyEvent(event schema.Event) error {
	switch e := event.(type) {
	case *schema.DeploymentCreatedEvent:
		return handleDeploymentCreatedEvent(r, e)
	case *schema.DeploymentSchemaUpdatedEvent:
		return handleDeploymentSchemaUpdatedEvent(r, e)
	case *schema.DeploymentReplicasUpdatedEvent:
		return handleDeploymentReplicasUpdatedEvent(r, e)
	case *schema.DeploymentActivatedEvent:
		return handleDeploymentActivatedEvent(r, e)
	case *schema.DeploymentDeactivatedEvent:
		return handleDeploymentDeactivatedEvent(r, e)
	case *schema.VerbRuntimeEvent:
		return handleVerbRuntimeEvent(r, e)
	case *schema.TopicRuntimeEvent:
		return handleTopicRuntimeEvent(r, e)
	case *schema.DatabaseRuntimeEvent:
		return handleDatabaseRuntimeEvent(r, e)
	case *schema.ModuleRuntimeEvent:
		return handleModuleRuntimeEvent(r, e)
	case *schema.ProvisioningCreatedEvent:
		return handleProvisioningCreatedEvent(r, e)
	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
}

func handleDeploymentCreatedEvent(t SchemaState, e *schema.DeploymentCreatedEvent) error {
	if existing := t.deployments[e.Key]; existing != nil {
		return nil
	}
	t.deployments[e.Key] = e.Schema
	return nil
}

func handleDeploymentSchemaUpdatedEvent(t SchemaState, e *schema.DeploymentSchemaUpdatedEvent) error {
	_, ok := t.deployments[e.Key]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)
	}
	t.deployments[e.Key] = e.Schema
	return nil
}

func handleDeploymentReplicasUpdatedEvent(t SchemaState, e *schema.DeploymentReplicasUpdatedEvent) error {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)
	}
	existing.ModRuntime().ModScaling().MinReplicas = int32(e.Replicas)
	return nil
}

func handleDeploymentActivatedEvent(t SchemaState, e *schema.DeploymentActivatedEvent) error {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)

	}
	existing.ModRuntime().ModDeployment().ActivatedAt = optional.Some(e.ActivatedAt)
	existing.ModRuntime().ModScaling().MinReplicas = int32(e.MinReplicas)
	t.activeDeployments[e.Key] = true
	return nil
}

func handleDeploymentDeactivatedEvent(t SchemaState, e *schema.DeploymentDeactivatedEvent) error {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)

	}
	existing.ModRuntime().ModScaling().MinReplicas = 0
	delete(t.activeDeployments, e.Key)
	return nil
}

func handleVerbRuntimeEvent(t SchemaState, e *schema.VerbRuntimeEvent) error {
	m, ok := t.provisioning[e.Module]
	if !ok {
		return fmt.Errorf("module %s not found", e.Module)
	}
	for verb := range slices.FilterVariants[*schema.Verb](m.Decls) {
		if verb.Name == e.ID {
			if verb.Runtime == nil {
				verb.Runtime = &schema.VerbRuntime{}
			}

			if base, ok := e.Base.Get(); ok {
				verb.Runtime.Base = base
			}
			if subscription, ok := e.Subscription.Get(); ok {
				verb.Runtime.Subscription = &subscription
			}
		}
	}
	return nil
}

func handleTopicRuntimeEvent(t SchemaState, e *schema.TopicRuntimeEvent) error {
	m, ok := t.provisioning[e.Module]
	if !ok {
		return fmt.Errorf("module %s not found", e.Module)
	}
	for topic := range slices.FilterVariants[*schema.Topic](m.Decls) {
		if topic.Name == e.ID {
			topic.Runtime = e.Payload
		}
	}
	return nil
}

func handleDatabaseRuntimeEvent(t SchemaState, e *schema.DatabaseRuntimeEvent) error {
	m, ok := t.provisioning[e.Module]
	if !ok {
		return fmt.Errorf("module %s not found", e.Module)
	}
	for _, decl := range m.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Name == e.ID {
			if db.Runtime == nil {
				db.Runtime = &schema.DatabaseRuntime{}
			}
			db.Runtime.Connections = e.Connections
		}
	}
	return nil
}

func handleModuleRuntimeEvent(t SchemaState, e *schema.ModuleRuntimeEvent) error {
	var module *schema.Module
	if dk, ok := e.DeploymentKey.Get(); ok {
		deployment, err := key.ParseDeploymentKey(dk)
		if err != nil {
			return fmt.Errorf("invalid deployment key: %w", err)
		}
		module = t.deployments[deployment]
	} else {
		// updating a provisioning module
		m, ok := t.provisioning[e.Module]
		if !ok {
			return fmt.Errorf("module %s not found", e.Module)
		}
		module = m
	}
	if base, ok := e.Base.Get(); ok {
		module.ModRuntime().Base = base
	}
	if scaling, ok := e.Scaling.Get(); ok {
		module.ModRuntime().Scaling = &scaling
	}
	if deployment, ok := e.Deployment.Get(); ok {
		module.ModRuntime().Deployment = &deployment
	}
	return nil
}

func handleProvisioningCreatedEvent(t SchemaState, e *schema.ProvisioningCreatedEvent) error {
	t.provisioning[e.DesiredModule.Name] = e.DesiredModule
	return nil
}
