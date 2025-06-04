package schemaservice

import (
	"context"
	goslices "slices"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"

	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
)

// TODO: these should be event methods once we can move them to this package

// ApplyEvent applies an event to the schema state
func (r *SchemaState) ApplyEvent(ctx context.Context, event schema.Event) error {
	logger := log.FromContext(ctx).Scope("schemaevents")
	logger.Debugf("Applying %s", event.DebugString())
	if err := event.Validate(); err != nil {
		return errors.Wrap(err, "invalid event")
	}
	switch e := event.(type) {
	case *schema.DeploymentRuntimeEvent:
		return errors.WithStack(handleDeploymentRuntimeEvent(r, e))
	case *schema.ChangesetCreatedEvent:
		return errors.WithStack(handleChangesetCreatedEvent(r, e))
	case *schema.ChangesetPreparedEvent:
		return errors.WithStack(handleChangesetPreparedEvent(r, e))
	case *schema.ChangesetCommittedEvent:
		return errors.WithStack(handleChangesetCommittedEvent(ctx, r, e))
	case *schema.ChangesetDrainedEvent:
		return errors.WithStack(handleChangesetDrainedEvent(ctx, r, e))
	case *schema.ChangesetFinalizedEvent:
		return errors.WithStack(handleChangesetFinalizedEvent(ctx, r, e))
	case *schema.ChangesetRollingBackEvent:
		return errors.WithStack(handleChangesetRollingBackEvent(r, e))
	case *schema.ChangesetFailedEvent:
		return errors.WithStack(handleChangesetFailedEvent(r, e))
	default:
		return errors.Errorf("unknown event type: %T", e)
	}
}

// ApplyEvents applies a list of events to the schema state.
// events might be partially applied if an error is returned.
func (r *SchemaState) ApplyEvents(ctx context.Context, events ...schema.Event) error {
	for _, event := range events {
		if err := r.ApplyEvent(ctx, event); err != nil {
			return errors.Wrapf(err, "error applying event %T", event)
		}
	}
	return nil
}

// VerifyEvent verifies an event is valid for the given state, without applying it
func (r *SchemaState) VerifyEvent(ctx context.Context, event schema.Event) error {
	if err := event.Validate(); err != nil {
		return errors.Wrap(err, "invalid event")
	}
	switch e := event.(type) {
	case *schema.DeploymentRuntimeEvent:
		return errors.WithStack(verifyDeploymentRuntimeEvent(r, e))
	case *schema.ChangesetCreatedEvent:
		return errors.WithStack(verifyChangesetCreatedEvent(r, e))
	case *schema.ChangesetPreparedEvent:
		return errors.WithStack(verifyChangesetPreparedEvent(r, e))
	case *schema.ChangesetCommittedEvent:
		return errors.WithStack(verifyChangesetCommittedEvent(r, e))
	case *schema.ChangesetDrainedEvent:
		return errors.WithStack(verifyChangesetDrainedEvent(r, e))
	case *schema.ChangesetFinalizedEvent:
		return errors.WithStack(verifyChangesetFinalizedEvent(r, e))
	case *schema.ChangesetRollingBackEvent:
		return errors.WithStack(verifyChangesetRollingBackEvent(r, e))
	case *schema.ChangesetFailedEvent:
		return errors.WithStack(verifyChangesetFailedEvent(r, e))
	default:
		return errors.Errorf("unknown event type: %T", e)
	}
}

func verifyDeploymentRuntimeEvent(t *SchemaState, e *schema.DeploymentRuntimeEvent) error {
	if cs, ok := e.ChangesetKey().Get(); ok && !cs.IsZero() {
		changeset, ok := t.GetChangeset(cs).Get()
		if !ok {
			return errors.Errorf("changeset %s not found", cs.String())
		}
		for _, m := range changeset.InternalRealm().Modules {
			if m.Name == e.DeploymentKey().Payload.Module {
				return nil
			}
		}
	}
	for _, m := range t.GetDeployments() {
		if m.Runtime.Deployment.DeploymentKey == e.DeploymentKey() {
			return nil
		}
	}
	return errors.Errorf("deployment %s not found", e.DeploymentKey().String())
}

func handleDeploymentRuntimeEvent(t *SchemaState, e *schema.DeploymentRuntimeEvent) error {
	if err := verifyDeploymentRuntimeEvent(t, e); err != nil {
		return errors.WithStack(err)
	}
	dk := e.DeploymentKey()
	module := dk.Payload.Module
	realm := dk.Payload.Realm

	if cs, ok := e.ChangesetKey().Get(); ok && !cs.IsZero() {
		changeset, ok := t.GetChangeset(cs).Get()
		if !ok {
			return errors.Errorf("changeset %s not found", cs.String())
		}
		for _, rc := range changeset.RealmChanges {
			if rc.Name != realm {
				continue
			}
			for _, m := range rc.Modules {
				if m.Name == module {
					err := e.Payload.ApplyToModule(m)
					if err != nil {
						return errors.Wrapf(err, "error applying runtime event to module %s", module)
					}
					t.state.ChangesetEvents = append(t.state.ChangesetEvents, e)
					return nil
				}
			}
		}
	}
	for _, m := range t.state.Schema.Realms {
		if m.Name != realm {
			continue
		}
		for _, m := range m.Modules {
			if m.Runtime.Deployment.DeploymentKey == e.DeploymentKey() {
				err := e.Payload.ApplyToModule(m)
				if err != nil {
					return errors.Wrapf(err, "error applying runtime event to module %s", m)
				}
			}
			t.state.DeploymentEvents = append(t.state.DeploymentEvents, e)
			return nil
		}
	}
	return errors.Errorf("deployment %s not found", e.DeploymentKey().String())
}

func verifyChangesetCreatedEvent(t *SchemaState, e *schema.ChangesetCreatedEvent) error {
	if _, ok := t.GetChangeset(e.Changeset.Key).Get(); ok {
		return errors.Errorf("changeset %s already exists ", e.Changeset.Key)
	}

	// validate there is at most one internal realm, and it matches the name of the realm in the state
	hasInternalRealm := false
	for _, rc := range e.Changeset.RealmChanges {
		if rc.External {
			continue
		}
		if hasInternalRealm {
			return errors.Errorf("changeset can have at most one internal realm")
		}
		hasInternalRealm = true
		for _, sr := range t.state.Schema.Realms {
			if sr.External {
				continue
			}
			if sr.Name != rc.Name {
				return errors.Errorf("internal realm must be called %s, got %s", sr.Name, rc.Name)
			}
		}

		existingModules := map[string]key.Changeset{}
		updatingModules := map[string]*schema.Module{}

		for _, cs := range t.GetChangesets() {
			if cs.ModulesAreCanonical() {
				for _, chrc := range cs.RealmChanges {
					if chrc.Name != rc.Name {
						continue
					}
					for _, mod := range chrc.Modules {
						existingModules[mod.Name] = cs.Key
						updatingModules[mod.Name] = mod
					}
				}
			}
		}
		for _, mod := range rc.Modules {
			if cs, ok := existingModules[mod.Name]; ok {
				return errors.Errorf("module %s is already being updated in changeset %s", mod.Name, cs.String())
			}
			if mod.Runtime == nil {
				return errors.Errorf("module %s has no runtime", mod.Name)
			}
			if mod.Runtime.Deployment == nil {
				return errors.Errorf("module %s has no deployment", mod.Name)
			}
			if mod.Runtime.Deployment.DeploymentKey.IsZero() {
				return errors.Errorf("module %s has no deployment key", mod.Name)
			}
			if mod.Runtime.Deployment.State == schema.DeploymentStateUnspecified {
				mod.Runtime.Deployment.State = schema.DeploymentStateProvisioning
			}
			if mod.Runtime.Deployment.State != schema.DeploymentStateProvisioning {
				return errors.Errorf("deployment %s is not in correct state expected %v got %v", mod.Name, schema.DeploymentStateProvisioning, mod.Runtime.Deployment.State)
			}
		}
		rem := map[string]bool{}
		for _, mod := range rc.ToRemove {
			rem[mod] = true
		}

		updatedDep := map[string]*schema.Module{}
		if realm, ok := t.state.Schema.Realm(rc.Name).Get(); ok {
			updatedDep = reflect.DeepCopy(realm.ModulesByName())
		}
		for _, mod := range updatingModules {
			updatedDep[mod.Name] = mod
		}
		sch := &schema.Schema{Realms: []*schema.Realm{{Modules: maps.Values(updatedDep)}}} //nolint
		merged := latestSchema(sch, e.Changeset)
		for _, realm := range merged.InternalRealms() {
			realm.Modules = slices.Filter(realm.Modules, func(m *schema.Module) bool {
				if m.Builtin {
					return true
				}
				remove := rem[m.Runtime.Deployment.DeploymentKey.String()]
				if remove {
					delete(rem, m.Runtime.Deployment.DeploymentKey.String())
				}
				return !remove
			})
		}
		if len(rem) > 0 {
			return errors.Errorf("changeset has modules to remove that are not in the schema: %v", maps.Keys(rem))
		}
		problems := []error{}
		for _, mod := range merged.InternalModules() {
			_, err := schema.ValidateModuleInSchema(merged, optional.Some(mod))
			if err != nil {
				problems = append(problems, errors.Wrapf(err, "module %s is not valid", mod.Name))
			}
		}
		if len(problems) > 0 {
			return errors.Wrap(errors.Join(problems...), "changeset failed validation")
		}
	}

	return nil
}

func handleChangesetCreatedEvent(t *SchemaState, e *schema.ChangesetCreatedEvent) error {
	if err := verifyChangesetCreatedEvent(t, e); err != nil {
		return errors.WithStack(err)
	}
	for _, rc := range e.Changeset.RealmChanges {
		if rc.External {
			continue
		}
		for _, dep := range rc.Modules {
			if dep.Runtime.Scaling == nil {
				if existing, ok := t.state.Schema.Module(rc.Name, dep.Name).Get(); ok {
					dep.Runtime.ModScaling().MinReplicas = existing.Runtime.Scaling.MinReplicas
				} else {
					dep.Runtime.Scaling = &schema.ModuleRuntimeScaling{MinReplicas: 1}
				}
			}
		}
	}
	t.upsertChangeset(e.Changeset)
	return nil
}

func verifyChangesetPreparedEvent(t *SchemaState, e *schema.ChangesetPreparedEvent) error {
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	for _, dep := range changeset.InternalRealm().Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateReady {
			return errors.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateReady, dep.Runtime.Deployment.State)
		}
		if !dep.ModRuntime().ModRunner().Provisioned() {
			return errors.Errorf("deployment %s has no endpoint, and an endpoint is required", dep.Name)
		}
	}
	return nil
}

func handleChangesetPreparedEvent(t *SchemaState, e *schema.ChangesetPreparedEvent) error {
	if err := verifyChangesetPreparedEvent(t, e); err != nil {
		return errors.WithStack(err)
	}
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStatePrepared
	// TODO: what does this actually mean? Worry about it when we start implementing canaries, but it will be clunky
	// If everything that cares about canaries needs to scan for prepared changesets
	for _, dep := range changeset.InternalRealm().Modules {
		dep.Runtime.Deployment.State = schema.DeploymentStateCanary
	}
	return nil
}

func verifyChangesetCommittedEvent(t *SchemaState, e *schema.ChangesetCommittedEvent) error {
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}

	for _, dep := range changeset.InternalRealm().Modules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateCanary {
			return errors.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateCanary, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetCommittedEvent(ctx context.Context, t *SchemaState, e *schema.ChangesetCommittedEvent) error {
	if err := verifyChangesetCommittedEvent(t, e); err != nil {
		return errors.WithStack(err)
	}

	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	logger := log.FromContext(ctx)
	changeset.State = schema.ChangesetStateCommitted
	for _, rc := range changeset.RealmChanges {
		realm, ok := t.state.Schema.Realm(rc.Name).Get()
		if !ok {
			realm = &schema.Realm{
				Name:     rc.Name,
				External: rc.External,
			}
			t.state.Schema.Realms = append(t.state.Schema.Realms, realm)
		}
		for _, dep := range rc.Modules {
			logger.Debugf("activating deployment %s %s", dep.GetRuntime().GetDeployment().DeploymentKey.String(), dep.Runtime.GetRunner().Endpoint)
			if old, ok := t.state.Schema.Module(rc.Name, dep.Name).Get(); ok {
				old.Runtime.Deployment.State = schema.DeploymentStateDraining
				rc.RemovingModules = append(rc.RemovingModules, old)
			}
			realm.UpsertModule(dep)
			t.clearDeploymentEvents(dep.Name)
			dep.Runtime.Deployment.State = schema.DeploymentStateCanonical
		}
		for _, dep := range rc.ToRemove {
			logger.Debugf("Removing deployment %s", dep)
			dk, err := key.ParseDeploymentKey(dep)
			if err != nil {
				logger.Errorf(err, "Error parsing deployment key %s", dep)
			} else {
				old, ok := t.state.Schema.Module(dk.Payload.Realm, dk.Payload.Module).Get()
				if !ok {
					return errors.Errorf("deployment %s not found", dk.Payload.Module)
				}
				old.Runtime.Deployment.State = schema.DeploymentStateDraining
				rc.RemovingModules = append(rc.RemovingModules, old)
				t.state.Schema.RemoveModule(dk.Payload.Realm, dk.Payload.Module)
			}
		}
	}
	return nil
}

func verifyChangesetDrainedEvent(t *SchemaState, e *schema.ChangesetDrainedEvent) error {
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateCommitted {
		return errors.Errorf("changeset %v is not in the correct state", changeset.Key)
	}

	for _, dep := range changeset.InternalRealm().RemovingModules {
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDraining &&
			dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeProvisioning {
			return errors.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateDeProvisioning, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetDrainedEvent(ctx context.Context, t *SchemaState, e *schema.ChangesetDrainedEvent) error {
	if err := verifyChangesetDrainedEvent(t, e); err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx)
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	logger.Debugf("Changeset %s drained", e.Key)

	for _, dep := range changeset.InternalRealm().RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDraining {
			dep.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
		}
	}
	changeset.State = schema.ChangesetStateDrained
	return nil
}
func verifyChangesetFinalizedEvent(t *SchemaState, e *schema.ChangesetFinalizedEvent) error {
	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	if changeset.State != schema.ChangesetStateDrained {
		return errors.Errorf("changeset %v is not in the correct state expected %v got %v", changeset.Key, schema.ChangesetStateDrained, changeset.State)
	}

	for _, dep := range changeset.InternalRealm().RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDeProvisioning {
			continue
		}
		if dep.ModRuntime().ModDeployment().State != schema.DeploymentStateDeleted {
			return errors.Errorf("deployment %s is not in correct state expected %v got %v", dep.Name, schema.DeploymentStateDeleted, dep.Runtime.Deployment.State)
		}
	}
	return nil
}

func handleChangesetFinalizedEvent(ctx context.Context, r *SchemaState, e *schema.ChangesetFinalizedEvent) error {
	if err := verifyChangesetFinalizedEvent(r, e); err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx)
	changeset, ok := r.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	logger.Debugf("Changeset %s de-provisioned", e.Key)

	for _, dep := range changeset.InternalRealm().RemovingModules {
		if dep.ModRuntime().ModDeployment().State == schema.DeploymentStateDeProvisioning {
			dep.Runtime.Deployment.State = schema.DeploymentStateDeleted
		}
	}
	changeset.State = schema.ChangesetStateFinalized
	// TODO: archive changesets?
	r.deleteChangeset(changeset.Key)
	// Archived changeset always has the most recent one at the head
	nl := []*schema.Changeset{changeset}
	nl = append(nl, r.archivedChangesets...)
	r.archivedChangesets = nl
	return nil
}

func verifyChangesetFailedEvent(t *SchemaState, e *schema.ChangesetFailedEvent) error {
	_, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	return nil
}

func handleChangesetFailedEvent(t *SchemaState, e *schema.ChangesetFailedEvent) error {
	if err := verifyChangesetFailedEvent(t, e); err != nil {
		return errors.WithStack(err)
	}

	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateFailed
	//TODO: de-provisioning on failure?
	t.deleteChangeset(changeset.Key)
	// Archived changeset always has the most recent one at the head
	nl := []*schema.Changeset{changeset}
	nl = append(nl, t.archivedChangesets...)
	t.archivedChangesets = nl
	return nil
}
func verifyChangesetRollingBackEvent(t *SchemaState, e *schema.ChangesetRollingBackEvent) error {
	cs, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	if cs.State != schema.ChangesetStatePrepared && cs.State != schema.ChangesetStatePreparing {
		return errors.Errorf("changeset %s is not in the correct state expected %v or %v got %v", cs.Key, schema.ChangesetStatePrepared, schema.ChangesetStatePreparing, cs.State)
	}
	return nil
}

func handleChangesetRollingBackEvent(t *SchemaState, e *schema.ChangesetRollingBackEvent) error {
	if err := verifyChangesetRollingBackEvent(t, e); err != nil {
		return errors.WithStack(err)
	}

	changeset, ok := t.GetChangeset(e.Key).Get()
	if !ok {
		return errors.Errorf("changeset %s not found", e.Key)
	}
	changeset.State = schema.ChangesetStateRollingBack
	changeset.Error = e.Error
	for _, module := range changeset.InternalRealm().Modules {
		module.Runtime.Deployment.State = schema.DeploymentStateDeProvisioning
	}

	return nil
}

// latest schema calculates the latest schema by applying active deployments in changeset to the canonical schema.
func latestSchema(canonical *schema.Schema, changeset *schema.Changeset) *schema.Schema {
	sch := reflect.DeepCopy(canonical)
	for _, realm := range sch.InternalRealms() {
		for _, module := range changeset.InternalRealm().Modules {
			if i := goslices.IndexFunc(realm.Modules, func(m *schema.Module) bool { return m.Name == module.Name }); i != -1 {
				realm.Modules[i] = module
			} else {
				realm.Modules = append(realm.Modules, module)
			}
		}
	}
	return sch
}
