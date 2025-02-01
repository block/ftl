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
func (r SchemaState) ApplyEvent(ctx context.Context, event schema.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
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
	case *schema.ChangesetCreatedEvent:
		return handleChangesetCreatedEvent(r, e)
	case *schema.ChangesetPreparedEvent:
		return handleChangesetPreparedEvent(r, e)
	case *schema.ChangesetCommittedEvent:
		return handleChangesetCommittedEvent(ctx, r, e)
	case *schema.ChangesetFailedEvent:
		return handleChangesetFailedEvent(r, e)
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
	t.activeDeployments[existing.Name] = e.Key
	return nil
}

func handleDeploymentDeactivatedEvent(t SchemaState, e *schema.DeploymentDeactivatedEvent) error {
	existing, ok := t.deployments[e.Key]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)
	}
	existing.ModRuntime().ModScaling().MinReplicas = 0
	if t.activeDeployments[existing.Name] == e.Key {
		delete(t.activeDeployments, existing.Name)
	}
	return nil
}

func handleVerbRuntimeEvent(t SchemaState, e *schema.VerbRuntimeEvent) error {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return err
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

func provisioningModule(t *SchemaState, module string) (*schema.Module, error) {
	for _, cs := range t.changesets {
		for _, d := range cs.Deployments {
			if d.Payload.Module == module {
				m, ok := t.deployments[d]
				if !ok {
					return nil, fmt.Errorf("deployment %s not found", d)
				}
				return m, nil
			}
		}
	}
	return nil, fmt.Errorf("module not found %s", module)
}

func handleTopicRuntimeEvent(t SchemaState, e *schema.TopicRuntimeEvent) error {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return err
	}
	for topic := range slices.FilterVariants[*schema.Topic](m.Decls) {
		if topic.Name == e.ID {
			topic.Runtime = e.Payload
			return nil
		}
	}
	return fmt.Errorf("topic %s not found", e.ID)
}

func handleDatabaseRuntimeEvent(t SchemaState, e *schema.DatabaseRuntimeEvent) error {
	m, err := provisioningModule(&t, e.Module)
	if err != nil {
		return err
	}
	for _, decl := range m.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Name == e.ID {
			if db.Runtime == nil {
				db.Runtime = &schema.DatabaseRuntime{}
			}
			db.Runtime.Connections = e.Connections
			return nil
		}
	}
	return fmt.Errorf("database %s not found", e.ID)
}

func handleModuleRuntimeEvent(t SchemaState, e *schema.ModuleRuntimeEvent) error {
	module := t.deployments[e.DeploymentKey]
	if module == nil {
		return fmt.Errorf("deployment %s not found", e.Deployment)
	}

	if base, ok := e.Base.Get(); ok {
		module.ModRuntime().Base = base
	}
	if scaling, ok := e.Scaling.Get(); ok {
		module.ModRuntime().Scaling = &scaling
	}
	if deployment, ok := e.Deployment.Get(); ok {
		if deployment.State == schema.DeploymentStateUnspecified {
			deployment.State = module.Runtime.Deployment.State
		} else if deployment.State != module.Runtime.Deployment.State {
			// We only allow a few different externally driven state changes
			if !(module.Runtime.Deployment.State == schema.DeploymentStateProvisioning && deployment.State == schema.DeploymentStateReady) {
				return fmt.Errorf("invalid state transition from %d to %d", module.Runtime.Deployment.State, deployment.State)
			}
		}
		module.ModRuntime().Deployment = &deployment
	}
	return nil
}

func handleChangesetCreatedEvent(t SchemaState, e *schema.ChangesetCreatedEvent) error {
	if existing := t.changesets[e.Changeset.Key]; existing != nil {
		return nil
	}
	if e.Changeset.State == schema.ChangesetStatePreparing {
		if active, ok := t.ActiveChangeset().Get(); ok {
			// TODO: make unit test for this
			// TODO: how does error handling work here? Does the changeset need to be added but immediately failed? Or is this error propagated to the caller?
			return fmt.Errorf("can not create active changeset: %s already active", active.Key)
		}
	}
	deployments := []key.Deployment{}
	for _, mod := range e.Changeset.Modules {
		if mod.Runtime == nil {
			return fmt.Errorf("module %s has no runtime", mod.Name)
		}
		if mod.Runtime.Deployment == nil {
			return fmt.Errorf("module %s has no deployment", mod.Name)
		}
		if mod.Runtime.Deployment.DeploymentKey.IsZero() {
			return fmt.Errorf("module %s has no deployment key", mod.Name)
		}
		if mod.Runtime.Deployment.State == schema.DeploymentStateUnspecified {
			mod.Runtime.Deployment.State = schema.DeploymentStateProvisioning
		}
		if mod.Runtime.Deployment.State != schema.DeploymentStateProvisioning {
			return fmt.Errorf("module %s is not in correct state", mod.Name)
		}
		deploymentKey := mod.Runtime.Deployment.DeploymentKey
		deployments = append(deployments, deploymentKey)
		t.deployments[deploymentKey] = mod
	}
	t.changesets[e.Changeset.Key] = &ChangesetDetails{
		Key:         e.Changeset.Key,
		CreatedAt:   e.Changeset.CreatedAt,
		Deployments: deployments,
		State:       e.Changeset.State,
		Error:       e.Changeset.Error,
	}
	return nil
}

func handleChangesetPreparedEvent(t SchemaState, e *schema.ChangesetPreparedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	for _, depName := range changeset.Deployments {
		dep := t.deployments[depName]
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateReady {
			return fmt.Errorf("deployment %s is not in correct state %d", depName, dep.ModRuntime().ModDeployment().State)
		}
	}
	changeset.State = schema.ChangesetStatePrepared
	// TODO: what does this actually mean? Worry about it when we start implementing canaries, but it will be clunky
	// If everything that cares about canaries needs to scan for prepared changesets
	for _, depName := range changeset.Deployments {
		t.deployments[depName].Runtime.Deployment.State = schema.DeploymentStateCanary
	}
	return nil
}

func handleChangesetCommittedEvent(ctx context.Context, t SchemaState, e *schema.ChangesetCommittedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}

	for _, depName := range changeset.Deployments {
		dep := t.deployments[depName]
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateCanary {
			return fmt.Errorf("deployment %s is not in correct state %d", depName, dep.ModRuntime().ModDeployment().State)
		}
	}
	logger := log.FromContext(ctx)
	changeset.State = schema.ChangesetStateCommitted
	for _, depName := range changeset.Deployments {
		logger.Debugf("activating deployment %s", t.deployments[depName].GetRuntime().GetDeployment().Endpoint)
		dep := t.deployments[depName]
		t.activeDeployments[dep.Name] = depName
		// We need proper cleanup
	}
	return nil
}

func handleChangesetFailedEvent(t SchemaState, e *schema.ChangesetFailedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateFailed
	changeset.Error = e.Error
	return nil
}
