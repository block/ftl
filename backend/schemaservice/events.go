package schemaservice

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// TODO: these should be event methods once we can move them to this package

// ApplyEvent applies an event to the schema state
func (r SchemaState) ApplyEvent(ctx context.Context, event schema.Event) (SchemaState, error) {
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
	case *schema.ChangesetCreatedEvent:
		return handleChangesetCreatedEvent(r, e)
	case *schema.ChangesetCommittedEvent:
		return handleChangesetCommittedEvent(ctx, r, e)
	case *schema.ChangesetFailedEvent:
		return handleChangesetFailedEvent(r, e)
	default:
		return r, fmt.Errorf("unknown event type: %T", e)
	}
}

func handleDeploymentCreatedEvent(t SchemaState, e *schema.DeploymentCreatedEvent) (SchemaState, error) {
	if existing := t.deployments[e.Key]; existing != nil {
		return t, nil
	}
	t.deployments[e.Key] = e.Schema
	return t, nil
}

func handleDeploymentSchemaUpdatedEvent(t SchemaState, e *schema.DeploymentSchemaUpdatedEvent) (SchemaState, error) {
	_, ok := t.deployments[e.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", e.Key)
	}
	t.deployments[e.Key] = e.Schema
	return t, nil
}

func handleDeploymentReplicasUpdatedEvent(t SchemaState, e *schema.DeploymentReplicasUpdatedEvent) (SchemaState, error) {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", e.Key)
	}
	existing.ModRuntime().ModScaling().MinReplicas = int32(e.Replicas)
	return t, nil
}

func handleDeploymentActivatedEvent(t SchemaState, e *schema.DeploymentActivatedEvent) (SchemaState, error) {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", e.Key)

	}
	existing.ModRuntime().ModDeployment().ActivatedAt = optional.Some(e.ActivatedAt)
	existing.ModRuntime().ModScaling().MinReplicas = int32(e.MinReplicas)
	t.activeDeployments[existing.Name] = e.Key
	return t, nil
}

func handleDeploymentDeactivatedEvent(t SchemaState, e *schema.DeploymentDeactivatedEvent) (SchemaState, error) {
	//TODO: fix deactivation
	existing, ok := t.deployments[e.Key]
	if !ok {
		return t, fmt.Errorf("deployment %s not found", e.Key)
	}
	existing.ModRuntime().ModScaling().MinReplicas = 0
	delete(t.activeDeployments, existing.Name)
	return t, nil
}

func handleVerbRuntimeEvent(t SchemaState, e *schema.VerbRuntimeEvent) (SchemaState, error) {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return SchemaState{}, err
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
	return t, nil
}

func provisioningModule(t *SchemaState, module string) (*schema.Module, error) {
	d, ok := t.provisioning[module]
	if !ok {
		return nil, fmt.Errorf("module %s not found", module)
	}
	m, ok := t.deployments[d]
	if !ok {
		return nil, fmt.Errorf("deployment %s not found", d)
	}
	return m, nil
}

func handleTopicRuntimeEvent(t SchemaState, e *schema.TopicRuntimeEvent) (SchemaState, error) {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return SchemaState{}, err
	}
	for topic := range slices.FilterVariants[*schema.Topic](m.Decls) {
		if topic.Name == e.ID {
			topic.Runtime = e.Payload
			return t, nil
		}
	}
	return t, fmt.Errorf("topic %s not found", e.ID)
}

func handleDatabaseRuntimeEvent(t SchemaState, e *schema.DatabaseRuntimeEvent) (SchemaState, error) {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return SchemaState{}, err
	}
	for _, decl := range m.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Name == e.ID {
			if db.Runtime == nil {
				db.Runtime = &schema.DatabaseRuntime{}
			}
			db.Runtime.Connections = e.Connections
			return t, nil
		}
	}
	return t, fmt.Errorf("database %s not found", e.ID)
}

func handleModuleRuntimeEvent(t SchemaState, e *schema.ModuleRuntimeEvent) (SchemaState, error) {
	module := t.deployments[e.DeploymentKey]
	if module == nil {
		return t, fmt.Errorf("deployment %s not found", e.Deployment)
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
	return t, nil
}

func handleProvisioningCreatedEvent(t SchemaState, e *schema.ProvisioningCreatedEvent) (SchemaState, error) {
	t.deployments[e.DesiredModule.Runtime.Deployment.DeploymentKey] = e.DesiredModule
	t.provisioning[e.DesiredModule.Name] = e.DesiredModule.Runtime.Deployment.DeploymentKey
	return t, nil
}

func handleChangesetCreatedEvent(t SchemaState, e *schema.ChangesetCreatedEvent) (SchemaState, error) {
	if existing := t.changesets[e.Changeset.Key]; existing != nil {
		return t, nil
	}
	if e.Changeset.State == schema.ChangesetStateProvisioning {
		if active, ok := t.ActiveChangeset().Get(); ok {
			// TODO: make unit test for this
			// TODO: how does error handling work here? Does the changeset need to be added but immediately failed? Or is this error propagated to the caller?
			return t, fmt.Errorf("can not create active changeset: %s already active", active.Key)
		}
	}
	deployments := []key.Deployment{}
	for _, mod := range e.Changeset.Modules {
		if mod.Runtime == nil {
			return t, fmt.Errorf("module %s has no runtime", mod.Name)
		}
		if mod.Runtime.Deployment == nil {
			return t, fmt.Errorf("module %s has no deployment", mod.Name)
		}
		if mod.Runtime.Deployment.DeploymentKey.IsZero() {
			return t, fmt.Errorf("module %s has no deployment key", mod.Name)
		}
		deploymentKey := mod.Runtime.Deployment.DeploymentKey
		deployments = append(deployments, deploymentKey)
		t.deployments[deploymentKey] = mod
		t.provisioning[mod.Name] = deploymentKey
	}
	t.changesets[e.Changeset.Key] = &ChangesetDetails{
		Key:         e.Changeset.Key,
		CreatedAt:   e.Changeset.CreatedAt,
		Deployments: deployments,
		State:       e.Changeset.State,
		Error:       e.Changeset.Error,
	}
	return t, nil
}

func handleChangesetCommittedEvent(ctx context.Context, t SchemaState, e *schema.ChangesetCommittedEvent) (SchemaState, error) {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return SchemaState{}, fmt.Errorf("changeset %s not found", e.Key)
	}
	logger := log.FromContext(ctx)
	changeset.State = schema.ChangesetStateCommitted
	for _, depName := range changeset.Deployments {
		logger.Infof("activating deployment %s", t.deployments[depName].GetRuntime().GetDeployment().Endpoint)
		dep := t.deployments[depName]
		t.activeDeployments[dep.Name] = depName
	}
	return t, nil
}

func handleChangesetFailedEvent(t SchemaState, e *schema.ChangesetFailedEvent) (SchemaState, error) {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return SchemaState{}, fmt.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateFailed
	changeset.Error = e.Error
	return t, nil
}
