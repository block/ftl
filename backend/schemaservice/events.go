package schemaservice

import (
	"context"
	"fmt"
	slices2 "slices"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/errors"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// TODO: these should be event methods once we can move them to this package

// ApplyEvent applies an event to the schema state
func (r SchemaState) ApplyEvent(ctx context.Context, event schema.Event) error {
	logger := log.FromContext(ctx)
	logger.Debugf("Applying event %s", event.DebugString())
	if err := event.Validate(); err != nil {
		logger.Errorf(err, "Failed to Apply Event")
		return nil
	}
	// TODO: Properly validate events before applying them to raft log.
	// If we return an error here after it is applied to the log, the raft host will crash.
	var err error
	switch e := event.(type) {
	case *schema.DeploymentSchemaUpdatedEvent:
		err = handleDeploymentSchemaUpdatedEvent(r, e)
	case *schema.DeploymentReplicasUpdatedEvent:
		err = handleDeploymentReplicasUpdatedEvent(r, e)
	case *schema.VerbRuntimeEvent:
		err = handleVerbRuntimeEvent(r, e)
	case *schema.TopicRuntimeEvent:
		err = handleTopicRuntimeEvent(r, e)
	case *schema.DatabaseRuntimeEvent:
		err = handleDatabaseRuntimeEvent(r, e)
	case *schema.ModuleRuntimeEvent:
		err = handleModuleRuntimeEvent(ctx, r, e)
	case *schema.ChangesetCreatedEvent:
		err = handleChangesetCreatedEvent(r, e)
	case *schema.ChangesetPreparedEvent:
		err = handleChangesetPreparedEvent(r, e)
	case *schema.ChangesetCommittedEvent:
		err = handleChangesetCommittedEvent(ctx, r, e)
	case *schema.ChangesetDrainedEvent:
		err = handleChangesetDrainedEvent(ctx, r, e)
	case *schema.ChangesetFinalizedEvent:
		err = handleChangesetFinalizedEvent(ctx, r, e)
	case *schema.ChangesetFailedEvent:
		err = handleChangesetFailedEvent(r, e)
	default:
		err = fmt.Errorf("unknown event type: %T", e)
	}
	if err != nil {
		logger.Errorf(err, "Failed to Handle Event")
	}
	return nil
}

func handleDeploymentSchemaUpdatedEvent(t SchemaState, e *schema.DeploymentSchemaUpdatedEvent) error {
	if e.Changeset.IsZero() {
		// The only reason this is optional is for event extract reasons that should change when it is refactored
		return fmt.Errorf("changeset is required")
	}
	cs, ok := t.changesets[e.Changeset]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	for i, m := range cs.Modules {
		if m.Name == e.Schema.Name {
			cs.Modules[i] = e.Schema
			return nil
		}
	}

	return fmt.Errorf("module %s not found in changeset %s", e.Schema.Name, e.Changeset)
}

func handleDeploymentReplicasUpdatedEvent(t SchemaState, e *schema.DeploymentReplicasUpdatedEvent) error {
	existing, ok := t.deployments[e.Key.Payload.Module]
	if !ok {
		return fmt.Errorf("deployment %s not found", e.Key)
	}
	existing.ModRuntime().ModScaling().MinReplicas = int32(e.Replicas)
	return nil
}

func handleVerbRuntimeEvent(t SchemaState, e *schema.VerbRuntimeEvent) error {
	m, storeEvent, err := t.handleRuntimeEvent(e)
	if err != nil {
		return err
	}
	for verb := range slices.FilterVariants[*schema.Verb](m.Decls) {
		if verb.Name == e.ID {
			if verb.Runtime == nil {
				verb.Runtime = &schema.VerbRuntime{}
			}

			if subscription, ok := e.Subscription.Get(); ok {
				verb.Runtime.Subscription = &subscription
			}
		}
	}
	storeEvent()
	return nil
}

func handleTopicRuntimeEvent(t SchemaState, e *schema.TopicRuntimeEvent) error {
	m, storeEvent, err := t.handleRuntimeEvent(e)
	if err != nil {
		return err
	}
	for topic := range slices.FilterVariants[*schema.Topic](m.Decls) {
		if topic.Name == e.ID {
			topic.Runtime = e.Payload
			storeEvent()
			return nil
		}
	}
	return fmt.Errorf("topic %s not found", e.ID)
}

func handleDatabaseRuntimeEvent(t SchemaState, e *schema.DatabaseRuntimeEvent) error {
	m, storeEvent, err := t.handleRuntimeEvent(e)
	if err != nil {
		return err
	}
	for _, decl := range m.Decls {
		if db, ok := decl.(*schema.Database); ok && db.Name == e.ID {
			if db.Runtime == nil {
				db.Runtime = &schema.DatabaseRuntime{}
			}
			db.Runtime.Connections = e.Connections
			storeEvent()
			return nil
		}
	}
	return fmt.Errorf("database %s not found", e.ID)
}

func handleModuleRuntimeEvent(ctx context.Context, t SchemaState, e *schema.ModuleRuntimeEvent) error {
	module, storeEvent, err := t.handleRuntimeEvent(e)
	if err != nil {
		return err
	}

	if base, ok := e.Base.Get(); ok {
		module.ModRuntime().Base = base
	}
	if scaling, ok := e.Scaling.Get(); ok {
		module.ModRuntime().Scaling = &scaling
	}
	if runner, ok := e.Runner.Get(); ok {
		module.ModRuntime().Runner = &runner
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
		log.FromContext(ctx).Debugf("deployment %s state change %v -> %v", module.Name, module.Runtime.Deployment, deployment)
		module.ModRuntime().Deployment = &deployment
	}
	storeEvent()
	return nil
}

func handleChangesetCreatedEvent(t SchemaState, e *schema.ChangesetCreatedEvent) error {
	if existing := t.changesets[e.Changeset.Key]; existing != nil {
		return fmt.Errorf("changeset %s already exists ", e.Changeset.Key)
	}
	activeCount := 0
	existingModules := map[string]key.Changeset{}
	for _, cs := range t.changesets {
		if cs.ModulesAreCanonical() {
			//TODO: at the moment changesets accumulate forever...
			for _, mod := range cs.Modules {
				existingModules[mod.Name] = cs.Key
			}
			activeCount++
		}
	}
	for _, mod := range e.Changeset.Modules {
		if cs, ok := existingModules[mod.Name]; ok {
			return fmt.Errorf("module %s is already being updated in changeset %s", mod.Name, cs.String())
		}
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
	}
	if activeCount > 0 {
		return fmt.Errorf("only a single changeset can currently be active at any time")
	}

	if t.validationEnabled {
		sch := &schema.Schema{Modules: maps.Values(t.deployments)}
		merged := latestSchema(sch, e.Changeset)
		problems := []error{}
		for _, mod := range e.Changeset.Modules {
			_, err := schema.ValidateModuleInSchema(merged, optional.Some(mod))
			if err != nil {
				problems = append(problems, fmt.Errorf("module %s is not valid: %w", mod.Name, err))
			}
		}
		if len(problems) > 0 {
			return fmt.Errorf("changest failed validation %w", errors.Join(problems...))
		}
	}
	t.changesets[e.Changeset.Key] = e.Changeset
	return nil
}

func handleChangesetPreparedEvent(t SchemaState, e *schema.ChangesetPreparedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	for _, dep := range changeset.Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateReady {
			return fmt.Errorf("deployment %s is not in correct state %d", dep.Name, dep.ModRuntime().ModDeployment().State)
		}
		if dep.ModRuntime().ModRunner().Endpoint == "" {
			return fmt.Errorf("deployment %s has no endpoint", dep.Name)
		}
	}
	changeset.State = schema.ChangesetStatePrepared
	// TODO: what does this actually mean? Worry about it when we start implementing canaries, but it will be clunky
	// If everything that cares about canaries needs to scan for prepared changesets
	for _, dep := range changeset.Modules {
		dep.Runtime.Deployment.State = schema.DeploymentStateCanary
	}
	return nil
}

func handleChangesetCommittedEvent(ctx context.Context, t SchemaState, e *schema.ChangesetCommittedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}

	for _, dep := range changeset.Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateCanary {
			return fmt.Errorf("deployment %s is not in correct state %d", dep.Name, dep.ModRuntime().ModDeployment().State)
		}
	}
	logger := log.FromContext(ctx)
	changeset.State = schema.ChangesetStateCommitted
	for _, dep := range changeset.Modules {
		logger.Debugf("activating deployment %s %s", dep.GetRuntime().GetDeployment().DeploymentKey.String(), dep.Runtime.GetRunner().Endpoint)
		if old, ok := t.deployments[dep.Name]; ok {
			old.Runtime.Deployment.State = schema.DeploymentStateDraining
			changeset.RemovingModules = append(changeset.RemovingModules, old)
		}
		t.deployments[dep.Name] = dep
		delete(t.deploymentEvents, dep.Name)
	}
	delete(t.changesetEvents, e.Key)
	return nil
}

func handleChangesetDrainedEvent(ctx context.Context, t SchemaState, e *schema.ChangesetDrainedEvent) error {
	logger := log.FromContext(ctx)
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateCommitted {
		return fmt.Errorf("changeset %v is not in the correct state", changeset.Key)
	}
	logger.Debugf("Changeset %s drained", e.Key)

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeProvisioning {
			return fmt.Errorf("deployment %s is not in correct state %d", dep.Name, dep.ModRuntime().ModDeployment().State)
		}
	}
	changeset.State = schema.ChangesetStateDrained
	return nil
}
func handleChangesetFinalizedEvent(ctx context.Context, t SchemaState, e *schema.ChangesetFinalizedEvent) error {
	logger := log.FromContext(ctx)
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateDrained {
		return fmt.Errorf("changeset %v is not in the correct state", changeset.Key)
	}
	logger.Debugf("Changeset %s de-provisioned", e.Key)

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeleted {
			return fmt.Errorf("deployment %s is not in correct state %d", dep.Name, dep.ModRuntime().ModDeployment().State)
		}
	}
	changeset.State = schema.ChangesetStateFinalized
	// TODO: archive changesets?
	delete(t.changesets, changeset.Key)
	t.archivedChangesets = append(t.archivedChangesets, changeset)
	return nil
}

func handleChangesetFailedEvent(t SchemaState, e *schema.ChangesetFailedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateFailed
	changeset.Error = e.Error
	//TODO: de-provisioning on failure?
	delete(t.changesets, changeset.Key)
	t.archivedChangesets = append(t.archivedChangesets, changeset)
	return nil
}

// latest schema calculates the latest schema by applying active deployments in changeset to the canonical schema.
func latestSchema(canonical *schema.Schema, changeset *schema.Changeset) *schema.Schema {
	sch := reflect.DeepCopy(canonical)
	for _, module := range changeset.Modules {
		if i := slices2.IndexFunc(sch.Modules, func(m *schema.Module) bool { return m.Name == module.Name }); i != -1 {
			sch.Modules[i] = module
		} else {
			sch.Modules = append(sch.Modules, module)
		}
	}
	return sch
}
