package schemaservice_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/schemaservice"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestDeploymentState(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cs := schemaservice.NewInMemorySchemaState(ctx)
	view, err := cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(view.GetDeployments()))

	deploymentKey := key.NewDeploymentKey("test-deployment")
	err = cs.Publish(ctx, schemaservice.EventWrapper{Event: &schema.DeploymentCreatedEvent{
		Key: deploymentKey,
		Schema: &schema.Module{
			Name: "test",
			Runtime: &schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{
					DeploymentKey: deploymentKey,
				},
			},
		},
	}})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(view.GetDeployments()))

	activate := time.Now()
	err = cs.Publish(ctx, schemaservice.EventWrapper{Event: &schema.DeploymentActivatedEvent{
		Key:         deploymentKey,
		ActivatedAt: activate,
		MinReplicas: 1,
	}})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, view.GetDeployments()[deploymentKey].GetRuntime().GetScaling().GetMinReplicas())
	assert.Equal(t, optional.Some(activate), view.GetDeployments()[deploymentKey].GetRuntime().GetDeployment().ActivatedAt)

	err = cs.Publish(ctx, schemaservice.EventWrapper{Event: &schema.DeploymentDeactivatedEvent{
		Key: deploymentKey,
	}})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, view.GetDeployments()[deploymentKey].GetRuntime().GetScaling().GetMinReplicas())
}
