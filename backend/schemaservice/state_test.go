package schemaservice

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestSchemaStateMarshalling(t *testing.T) {
	k := key.NewDeploymentKey("test")
	state := SchemaState{
		validationEnabled: true,
		deployments: map[string]*schema.Module{
			"test": {
				Name: "test",
				Runtime: &schema.ModuleRuntime{
					Deployment: &schema.ModuleRuntimeDeployment{
						DeploymentKey: k,
					},
				},
			},
		},
	}

	assertRoundTrip(t, state)
}

func TestStateMarshallingAfterCommonEvents(t *testing.T) {
	state := NewSchemaState(true)

	deploymentKey := key.NewDeploymentKey("test2")
	changesetKey := key.NewChangesetKey()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	assert.NoError(t, state.ApplyEvent(ctx, &schema.ChangesetCreatedEvent{
		Changeset: &schema.Changeset{
			Key: changesetKey,
			Modules: []*schema.Module{
				&schema.Module{
					Name: "test2",
					Runtime: &schema.ModuleRuntime{
						Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: deploymentKey},
					},
				},
			},
		},
	}))
	assert.NoError(t, state.ApplyEvent(ctx, &schema.DeploymentRuntimeEvent{
		Payload:   &schema.RuntimeElement{Deployment: deploymentKey, Element: &schema.ModuleRuntimeRunner{Endpoint: "http://localhost:8080"}},
		Changeset: &changesetKey,
	}))
	assert.NoError(t, state.ApplyEvent(ctx, &schema.DeploymentRuntimeEvent{
		Payload:   &schema.RuntimeElement{Deployment: deploymentKey, Element: &schema.ModuleRuntimeDeployment{State: schema.DeploymentStateReady}},
		Changeset: &changesetKey,
	}))
	assert.NoError(t, state.ApplyEvent(ctx, &schema.ChangesetPreparedEvent{
		Key: changesetKey,
		// No ActivatedAt, as proto conversion does not retain timezone
	}))
	assert.NoError(t, state.ApplyEvent(ctx, &schema.ChangesetCommittedEvent{
		Key: changesetKey,
		// No ActivatedAt, as proto conversion does not retain timezone
	}))
	assertRoundTrip(t, state)
}

func assertRoundTrip(t *testing.T, state SchemaState) {
	t.Helper()
	bytes, err := state.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal schema state: %v", err)
	}

	unmarshalledState := NewSchemaState(true)
	if err := unmarshalledState.Unmarshal(bytes); err != nil {
		t.Fatalf("failed to unmarshal schema state: %v", err)
	}
	assert.Equal(t, state, unmarshalledState)
}
