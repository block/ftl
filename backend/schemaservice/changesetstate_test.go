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
				DeploymentKey: key.NewDeploymentKey("test"),
			},
		},
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	state := schemaservice.NewInMemorySchemaState(ctx)
	view, err := state.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(view.GetDeployments()))
	assert.Equal(t, 0, len(view.GetChangesets()))

	t.Run("changeset must have id", func(t *testing.T) {
		err = state.Publish(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				CreatedAt: time.Now(),
				Modules:   []*schema.Module{module},
				Error:     "",
			},
		})
		assert.Error(t, err)
	})

	t.Run("deployment must must have deployment key", func(t *testing.T) {
		nm := reflect.DeepCopy(module)
		nm.Runtime = nil
		err = state.Publish(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:       key.NewChangesetKey(),
				CreatedAt: time.Now(),
				Modules:   []*schema.Module{nm},
				Error:     "",
			},
		})
		assert.Error(t, err)
	})

	var changesetKey key.Changeset
	t.Run("test create changeset", func(t *testing.T) {
		err = state.Publish(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key:       key.NewChangesetKey(),
				CreatedAt: time.Now(),
				Modules:   []*schema.Module{module},
				Error:     "",
			},
		})

		assert.NoError(t, err)
		view, err = state.View(ctx)
		assert.NoError(t, err)
		csd := changeset(t, view)
		assert.Equal(t, 1, len(csd.Modules))
		for _, d := range csd.Modules {
			assert.NoError(t, err)
			assert.Equal(t, schema.DeploymentStateProvisioning, d.Runtime.Deployment.State)
			changesetKey = csd.Key
		}
	})

	t.Run("test update module schema", func(t *testing.T) {
		newState := reflect.DeepCopy(module)
		newState.Runtime.Deployment.Endpoint = "http://localhost:8080"
		err = state.Publish(ctx, &schema.DeploymentSchemaUpdatedEvent{
			Key:       module.Runtime.Deployment.DeploymentKey,
			Schema:    newState,
			Changeset: changesetKey,
		})
		assert.NoError(t, err)
		view, err = state.View(ctx)
		assert.NoError(t, err)
		csd := changeset(t, view)
		assert.Equal(t, 1, len(csd.Modules))
		for _, d := range csd.Modules {
			assert.Equal(t, "http://localhost:8080", d.Runtime.Deployment.Endpoint)
			assert.Equal(t, schema.DeploymentStateProvisioning, d.Runtime.Deployment.State)
		}
	})

	t.Run("test commit changeset in bad state", func(t *testing.T) {
		// The deployment is not provisioned yet, this should fail
		err = state.Publish(ctx, &schema.ChangesetCommittedEvent{
			Key: changesetKey,
		})
		assert.Error(t, err)
	})

	t.Run("test prepare changeset in bad state", func(t *testing.T) {
		// The deployment is not provisioned yet, this should fail
		err = state.Publish(ctx, &schema.ChangesetPreparedEvent{
			Key: changesetKey,
		})
		assert.Error(t, err)
	})

	t.Run("test prepare changeset", func(t *testing.T) {
		newState := reflect.DeepCopy(module)
		newState.Runtime.Deployment.State = schema.DeploymentStateReady
		err = state.Publish(ctx, &schema.DeploymentSchemaUpdatedEvent{
			Key:       module.Runtime.Deployment.DeploymentKey,
			Schema:    newState,
			Changeset: changesetKey,
		})
		assert.NoError(t, err)
		view, err = state.View(ctx)
		assert.NoError(t, err)

		err = state.Publish(ctx, &schema.ChangesetPreparedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		view, err = state.View(ctx)
		assert.NoError(t, err)
		csd := changeset(t, view)
		assert.Equal(t, schema.ChangesetStatePrepared, csd.State)
		assert.Equal(t, 0, len(view.GetCanonicalDeployments()))
	})

	t.Run("test commit changeset", func(t *testing.T) {
		err = state.Publish(ctx, &schema.ChangesetCommittedEvent{
			Key: changesetKey,
		})
		assert.NoError(t, err)
		view, err = state.View(ctx)
		assert.NoError(t, err)
		csd := changeset(t, view)
		assert.Equal(t, schema.ChangesetStateCommitted, csd.State)
		assert.Equal(t, 1, len(view.GetCanonicalDeployments()))

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
