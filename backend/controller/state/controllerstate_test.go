package state_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/backend/controller/state"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestRunnerState(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	cs := state.NewInMemoryRunnerState(ctx)
	view, err := cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(view.Runners()))
	runnerkey := key.NewLocalRunnerKey(1)
	create := time.Now()
	endpoint := "http://localhost:8080"
	module := "test"
	deploymentKey := key.NewDeploymentKey(module)

	err = cs.Publish(ctx, &state.RunnerRegisteredEvent{
		Key:        runnerkey,
		Time:       create,
		Endpoint:   endpoint,
		Module:     module,
		Deployment: deploymentKey,
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(view.Runners()))
	assert.Equal(t, runnerkey, view.Runners()[0].Key)
	assert.Equal(t, create, view.Runners()[0].Create)
	assert.Equal(t, create, view.Runners()[0].LastSeen)
	assert.Equal(t, endpoint, view.Runners()[0].Endpoint)
	assert.Equal(t, module, view.Runners()[0].Module)
	assert.Equal(t, deploymentKey, view.Runners()[0].Deployment)
	seen := time.Now()
	err = cs.Publish(ctx, &state.RunnerRegisteredEvent{
		Key:        runnerkey,
		Time:       seen,
		Endpoint:   endpoint,
		Module:     module,
		Deployment: deploymentKey,
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, seen, view.Runners()[0].LastSeen)

	err = cs.Publish(ctx, &state.RunnerDeletedEvent{
		Key: runnerkey,
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(view.Runners()))

}

func TestDeploymentState(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cs := state.NewInMemorySchemaState(ctx)
	view, err := cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(view.GetDeployments()))

	deploymentKey := key.NewDeploymentKey("test-deployment")
	create := time.Now()
	err = cs.Publish(ctx, &state.DeploymentCreatedEvent{
		Key:       deploymentKey,
		CreatedAt: create,
		Schema:    &schema.Module{Name: "test"},
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(view.GetDeployments()))
	assert.Equal(t, create, view.GetDeployments()[deploymentKey].Schema.GetRuntime().GetDeployment().CreatedAt)

	activate := time.Now()
	err = cs.Publish(ctx, &state.DeploymentActivatedEvent{
		Key:         deploymentKey,
		ActivatedAt: activate,
		MinReplicas: 1,
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, view.GetDeployments()[deploymentKey].Schema.GetRuntime().GetScaling().GetMinReplicas())
	assert.Equal(t, activate, view.GetDeployments()[deploymentKey].Schema.GetRuntime().GetDeployment().ActivatedAt)

	err = cs.Publish(ctx, &state.DeploymentDeactivatedEvent{
		Key: deploymentKey,
	})
	assert.NoError(t, err)
	view, err = cs.View(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, view.GetDeployments()[deploymentKey].Schema.GetRuntime().GetScaling().GetMinReplicas())
}
