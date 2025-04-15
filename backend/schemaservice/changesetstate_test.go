package schemaservice_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/common/reflect"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestChangesetState(t *testing.T) {
	module := &schema.Module{
		Name: "test",
		Runtime: &schema.ModuleRuntime{
			Deployment: &schema.ModuleRuntimeDeployment{
				DeploymentKey: key.NewDeploymentKey("test", "test"),
			},
		},
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	state := schemaservice.NewSchemaState()

	assert.Equal(t, 0, len(state.GetDeployments()))
	assert.Equal(t, 0, len(state.GetChangesets()))

	t.Run("changeset must have id", func(t *testing.T) {
		event := &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				CreatedAt: time.Now(),
				RealmChanges: []*schema.RealmChange{{
					Name:    "test",
					Modules: []*schema.Module{module},
				}},
			},
		}
		sm := schemaservice.SchemaState{}
		assert.Error(t, sm.ApplyEvent(ctx, event))
	})

	t.Run("deployment must must have deployment key", func(t *testing.T) {
		nm := reflect.DeepCopy(module)
		nm.Runtime = nil
		event := &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:       key.NewChangesetKey(),
				CreatedAt: time.Now(),
				RealmChanges: []*schema.RealmChange{{
					Name:    "test",
					Modules: []*schema.Module{nm},
				}},
			},
		}
		sm := schemaservice.NewSchemaState()
		assert.Error(t, sm.ApplyEvent(ctx, event))
	})

	t.Run("changeset with two internal realms is rejected", func(t *testing.T) {
		event := &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:          key.NewChangesetKey(),
				CreatedAt:    time.Now(),
				RealmChanges: []*schema.RealmChange{{Name: "testrealm1"}, {Name: "testrealm2"}},
			},
		}

		sm := schemaservice.NewSchemaState()
		assert.Error(t, sm.ApplyEvent(ctx, event))
	})

	changesetKey := key.NewChangesetKey()
	t.Run("changeset with a new realm", func(t *testing.T) {
		err := state.ApplyEvent(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:       changesetKey,
				CreatedAt: time.Now(),
				RealmChanges: []*schema.RealmChange{{
					Name:    "test",
					Modules: []*schema.Module{module},
				}},
				Error: "",
			},
		})

		assert.NoError(t, err)
		csd := changeset(t, state)

		assert.Equal(t, 1, len(csd.RealmChanges))
		assert.Equal(t, 1, len(csd.RealmChanges[0].Modules))
		assert.Equal(t, "test", csd.RealmChanges[0].Name)

		for _, d := range csd.RealmChanges[0].Modules {
			assert.Equal(t, schema.DeploymentStateProvisioning, d.Runtime.Deployment.State)
			changesetKey = csd.Key
		}
	})

	t.Run("update module schema", func(t *testing.T) {
		newState := reflect.DeepCopy(module)
		newState.ModRuntime().ModRunner().Endpoint = "http://localhost:8080"
		err := state.ApplyEvent(ctx, &schema.DeploymentRuntimeEvent{
			Payload: &schema.RuntimeElement{
				Deployment: module.Runtime.Deployment.DeploymentKey,
				Element:    &schema.ModuleRuntimeRunner{Endpoint: "http://localhost:8080"},
			},
			Changeset: &changesetKey,
		})
		assert.NoError(t, err)
		csd := changeset(t, state)
		assert.Equal(t, 1, len(csd.InternalRealm().Modules))
		for _, d := range csd.InternalRealm().Modules {
			assert.Equal(t, "http://localhost:8080", d.Runtime.Runner.Endpoint)
			assert.Equal(t, schema.DeploymentStateProvisioning, d.Runtime.Deployment.State)
		}
	})

	t.Run("commit changeset in bad state", func(t *testing.T) {
		// The deployment is not provisioned yet, this should fail
		event := &schema.ChangesetCommittedEvent{
			Key: changesetKey,
		}
		err := state.ApplyEvent(ctx, event)
		assert.Error(t, err)
	})

	t.Run("prepare changeset in bad state", func(t *testing.T) {
		// The deployment is not provisioned yet, this should fail
		event := &schema.ChangesetPreparedEvent{
			Key: changesetKey,
		}
		err := state.ApplyEvent(ctx, event)
		assert.Error(t, err)
	})

	t.Run("prepare changeset", func(t *testing.T) {
		newState := reflect.DeepCopy(module)
		newState.Runtime.Deployment.State = schema.DeploymentStateReady
		err := state.ApplyEvent(ctx, &schema.DeploymentRuntimeEvent{
			Payload: &schema.RuntimeElement{
				Deployment: module.Runtime.Deployment.DeploymentKey,
				Element: &schema.ModuleRuntimeDeployment{
					State: schema.DeploymentStateReady,
				},
			},
			Changeset: &changesetKey,
		})
		assert.NoError(t, err)

		err = state.ApplyEvent(ctx, &schema.ChangesetPreparedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		csd := changeset(t, state)
		assert.Equal(t, schema.ChangesetStatePrepared, csd.State)
		assert.Equal(t, 0, len(state.GetCanonicalDeployments()))
	})

	t.Run("commit changeset", func(t *testing.T) {
		err := state.ApplyEvent(ctx, &schema.ChangesetCommittedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		csd := changeset(t, state)
		assert.Equal(t, schema.ChangesetStateCommitted, csd.State)
		assert.Equal(t, 1, len(state.GetCanonicalDeployments()))

		sch := state.GetCanonicalSchema()
		assert.Equal(t, 1, len(sch.Realms))
		assert.Equal(t, 1, len(sch.Realms[0].Modules))
	})

	t.Run("archive first changeset", func(t *testing.T) {
		err := state.ApplyEvent(ctx, &schema.ChangesetDrainedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		csd := changeset(t, state)
		assert.Equal(t, schema.ChangesetStateDrained, csd.State)
		assert.Equal(t, 1, len(state.GetChangesets()))

		err = state.ApplyEvent(ctx, &schema.ChangesetFinalizedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(state.GetChangesets()))
	})

	t.Run("changeset with a different intrnal realm is rejected", func(t *testing.T) {
		err := state.ApplyEvent(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:          changesetKey,
				RealmChanges: []*schema.RealmChange{{Name: "testrealm2"}},
				Error:        "",
			},
		})
		assert.Error(t, err)
	})
}

func changeset(t *testing.T, view schemaservice.SchemaState) *schema.Changeset {
	assert.Equal(t, 1, len(view.GetChangesets()))
	var csd *schema.Changeset
	for k := range view.GetChangesets() {
		csd = view.GetChangesets()[k]
		assert.Equal(t, k, csd.Key)
	}
	return csd
}
