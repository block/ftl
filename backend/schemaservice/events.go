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
func (r *SchemaState) ApplyEvent(ctx context.Context, event schema.Event) error {
	logger := log.FromContext(ctx).Scope("schemaevents")
	logger.Debugf("Applying %s", event.DebugString())
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	switch e := event.(type) {
	case *schema.DeploymentRuntimeEvent:
		return handleDeploymentRuntimeEvent(r, e)
	case *schema.ChangesetCreatedEvent:
		return handleChangesetCreatedEvent(r, e)
	case *schema.ChangesetPreparedEvent:
		return handleChangesetPreparedEvent(r, e)
	case *schema.ChangesetCommittedEvent:
		return handleChangesetCommittedEvent(ctx, r, e)
	case *schema.ChangesetDrainedEvent:
		return handleChangesetDrainedEvent(ctx, r, e)
	case *schema.ChangesetFinalizedEvent:
		return handleChangesetFinalizedEvent(ctx, r, e)
	case *schema.ChangesetRollingBackEvent:
		return handleChangesetRollingBackEvent(r, e)
	case *schema.ChangesetFailedEvent:
		return handleChangesetFailedEvent(r, e)
	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
}

// VerifyEvent verifies an event is valid for the given state, without applying it
func (r *SchemaState) VerifyEvent(ctx context.Context, event schema.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("invalid event: %w", err)
	}
	switch e := event.(type) {
	case *schema.DeploymentRuntimeEvent:
		return verifyDeploymentRuntimeEvent(r, e)
	case *schema.ChangesetCreatedEvent:
		return verifyChangesetCreatedEvent(r, e)
	case *schema.ChangesetPreparedEvent:
		return verifyChangesetPreparedEvent(r, e)
	case *schema.ChangesetCommittedEvent:
		return verifyChangesetCommittedEvent(r, e)
	case *schema.ChangesetDrainedEvent:
		return verifyChangesetDrainedEvent(r, e)
	case *schema.ChangesetFinalizedEvent:
		return verifyChangesetFinalizedEvent(r, e)
	case *schema.ChangesetRollingBackEvent:
		return verifyChangesetRollingBackEvent(r, e)
	case *schema.ChangesetFailedEvent:
		return verifyChangesetFailedEvent(r, e)
	default:
		return fmt.Errorf("unknown event type: %T", e)
	}
}

func verifyDeploymentRuntimeEvent(t *SchemaState, e *schema.DeploymentRuntimeEvent) error {
	if cs, ok := e.ChangesetKey().Get(); ok {
		_, ok := t.changesets[cs]
		if !ok {
			return fmt.Errorf("changeset %s not found", cs.String())
		}
		for _, m := range t.changesets[cs].Modules {
			if m.Name == e.DeploymentKey().Payload.Module {
				return nil
			}
		}
	}
	for _, m := range t.deployments {
		if m.Runtime.Deployment.DeploymentKey == e.DeploymentKey() {
			return nil
		}
	}
	return fmt.Errorf("deployment %s not found", e.DeploymentKey().String())
}

func handleDeploymentRuntimeEvent(t *SchemaState, e *schema.DeploymentRuntimeEvent) error {
	if err := verifyDeploymentRuntimeEvent(t, e); err != nil {
		return err
	}
	if cs, ok := e.ChangesetKey().Get(); ok {
		module := e.DeploymentKey().Payload.Module
		c := t.changesets[cs]
		for _, m := range c.Modules {
			if m.Name == module {
				err := e.Payload.ApplyToModule(m)
				if err != nil {
					return fmt.Errorf("error applying runtime event to module %s: %w", module, err)
				}
				t.changesetEvents[cs] = append(t.changesetEvents[cs], e)
				return nil
			}
		}
	}
	for k, m := range t.deployments {
		if m.Runtime.Deployment.DeploymentKey == e.DeploymentKey() {
			err := e.Payload.ApplyToModule(m)
			if err != nil {
				return fmt.Errorf("error applying runtime event to module %s: %w", m, err)
			}
			t.deploymentEvents[k] = append(t.deploymentEvents[k], e)
			return nil
		}
	}
	return fmt.Errorf("deployment %s not found", e.DeploymentKey().String())
}

func verifyChangesetCreatedEvent(t *SchemaState, e *schema.ChangesetCreatedEvent) error {
	if existing := t.changesets[e.Changeset.Key]; existing != nil {
		return fmt.Errorf("changeset %s already exists ", e.Changeset.Key)
	}
	activeCount := 0
	existingModules := map[string]key.Changeset{}
	for _, cs := range t.changesets {
		if cs.ModulesAreCanonical() {
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
			return fmt.Errorf("deployment %s is not in correct state expected %v got %v", mod.Name, schema.DeploymentStateProvisioning, mod.Runtime.Deployment.State)
		}
	}
	rem := map[string]bool{}
	for _, mod := range e.Changeset.ToRemove {
		rem[mod] = true
	}
	sch := &schema.Schema{Modules: maps.Values(t.deployments)}
	merged := latestSchema(sch, e.Changeset)
	merged.Modules = slices.Filter(merged.Modules, func(m *schema.Module) bool {
		if m.Builtin {
			return true
		}
		remove := rem[m.Runtime.Deployment.DeploymentKey.String()]
		if remove {
			delete(rem, m.Runtime.Deployment.DeploymentKey.String())
		}
		return !remove
	})
	if len(rem) > 0 {
		return fmt.Errorf("changeset has modules to remove that are not in the schema: %v", maps.Keys(rem))
	}
	problems := []error{}
	for _, mod := range merged.Modules {
		_, err := schema.ValidateModuleInSchema(merged, optional.Some(mod))
		if err != nil {
			problems = append(problems, fmt.Errorf("module %s is not valid: %w", mod.Name, err))
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("changeset failed validation %w", errors.Join(problems...))
	}

	return nil
}

func handleChangesetCreatedEvent(t *SchemaState, e *schema.ChangesetCreatedEvent) error {
	if err := verifyChangesetCreatedEvent(t, e); err != nil {
		return err
	}
	t.changesets[e.Changeset.Key] = e.Changeset
	return nil
}

func verifyChangesetPreparedEvent(t *SchemaState, e *schema.ChangesetPreparedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	for _, dep := range changeset.Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateReady {
			return fmt.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateReady, dep.Runtime.Deployment.State)
		}
		if dep.ModRuntime().ModRunner().Endpoint == "" {
			return fmt.Errorf("deployment %s has no endpoint", dep.Name)
		}
	}
	return nil
}

func handleChangesetPreparedEvent(t *SchemaState, e *schema.ChangesetPreparedEvent) error {
	if err := verifyChangesetPreparedEvent(t, e); err != nil {
		return err
	}
	changeset := t.changesets[e.Key]
	changeset.State = schema.ChangesetStatePrepared
	// TODO: what does this actually mean? Worry about it when we start implementing canaries, but it will be clunky
	// If everything that cares about canaries needs to scan for prepared changesets
	for _, dep := range changeset.Modules {
		dep.Runtime.Deployment.State = schema.DeploymentStateCanary
	}
	return nil
}

func verifyChangesetCommittedEvent(t *SchemaState, e *schema.ChangesetCommittedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}

	for _, dep := range changeset.Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateCanary {
			return fmt.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateCanary, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetCommittedEvent(ctx context.Context, t *SchemaState, e *schema.ChangesetCommittedEvent) error {
	if err := verifyChangesetCommittedEvent(t, e); err != nil {
		return err
	}

	changeset := t.changesets[e.Key]
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
		dep.Runtime.Deployment.State = schema.DeploymentStateCanonical
	}
	for _, dep := range changeset.ToRemove {
		logger.Debugf("Removing deployment %s", dep)
		dk, err := key.ParseDeploymentKey(dep)
		if err != nil {
			logger.Errorf(err, "Error parsing deployment key %s", dep)
		} else {
			old := t.deployments[dk.Payload.Module]
			old.Runtime.Deployment.State = schema.DeploymentStateDraining
			changeset.RemovingModules = append(changeset.RemovingModules, old)
			delete(t.deployments, dk.Payload.Module)
		}
	}
	return nil
}

func verifyChangesetDrainedEvent(t *SchemaState, e *schema.ChangesetDrainedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateCommitted {
		return fmt.Errorf("changeset %v is not in the correct state", changeset.Key)
	}

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDraining &&
			dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeProvisioning {
			return fmt.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateDeProvisioning, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetDrainedEvent(ctx context.Context, t *SchemaState, e *schema.ChangesetDrainedEvent) error {
	if err := verifyChangesetDrainedEvent(t, e); err != nil {
		return err
	}

	logger := log.FromContext(ctx)
	changeset := t.changesets[e.Key]
	logger.Debugf("Changeset %s drained", e.Key)

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDraining {
			dep.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
		}
	}
	changeset.State = schema.ChangesetStateDrained
	return nil
}
func verifyChangesetFinalizedEvent(t *SchemaState, e *schema.ChangesetFinalizedEvent) error {
	changeset, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateDrained {
		return fmt.Errorf("changeset %v is not in the correct state expected %v got %v", changeset.Key, schema.ChangesetStateDrained, changeset.State)
	}

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDeProvisioning {
			continue
		}
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeleted {
			return fmt.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateDeleted, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetFinalizedEvent(ctx context.Context, r *SchemaState, e *schema.ChangesetFinalizedEvent) error {
	if err := verifyChangesetFinalizedEvent(r, e); err != nil {
		return err
	}

	logger := log.FromContext(ctx)
	changeset := r.changesets[e.Key]
	logger.Debugf("Changeset %s de-provisioned", e.Key)

	for _, dep := range changeset.RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDeProvisioning {
			dep.Runtime.Deployment.State = schema.DeploymentStateDeleted
		}
	}
	changeset.State = schema.ChangesetStateFinalized
	// TODO: archive changesets?
	delete(r.changesets, changeset.Key)
	delete(r.changesetEvents, e.Key)
	// Archived changeset always has the most recent one at the head
	nl := []*schema.Changeset{changeset}
	nl = append(nl, r.archivedChangesets...)
	r.archivedChangesets = nl
	return nil
}

func verifyChangesetFailedEvent(t *SchemaState, e *schema.ChangesetFailedEvent) error {
	_, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	return nil
}

func handleChangesetFailedEvent(t *SchemaState, e *schema.ChangesetFailedEvent) error {
	if err := verifyChangesetFailedEvent(t, e); err != nil {
		return err
	}

	changeset := t.changesets[e.Key]
	changeset.State = schema.ChangesetStateFailed
	//TODO: de-provisioning on failure?
	delete(t.changesets, changeset.Key)
	// Archived changeset always has the most recent one at the head
	nl := []*schema.Changeset{changeset}
	nl = append(nl, t.archivedChangesets...)
	t.archivedChangesets = nl
	return nil
}
func verifyChangesetRollingBackEvent(t *SchemaState, e *schema.ChangesetRollingBackEvent) error {
	_, ok := t.changesets[e.Key]
	if !ok {
		return fmt.Errorf("changeset %s not found", e.Key)
	}
	return nil
}

func handleChangesetRollingBackEvent(t *SchemaState, e *schema.ChangesetRollingBackEvent) error {
	if err := verifyChangesetRollingBackEvent(t, e); err != nil {
		return err
	}

	changeset := t.changesets[e.Key]
	changeset.State = schema.ChangesetStateRollingBack
	changeset.Error = e.Error
	for _, module := range changeset.Modules {
		module.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
	}

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
