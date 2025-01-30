package schemaservice

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func TestSchemaStateMarshalling(t *testing.T) {
	k := key.NewDeploymentKey("test")
	state := SchemaState{
		deployments: map[key.Deployment]*schema.Module{
			k: {
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
	state := NewSchemaState()
	assert.NoError(t, state.ApplyEvent(&schema.ProvisioningCreatedEvent{
		DesiredModule: &schema.Module{Name: "test1"},
	}))
	deploymentKey := key.NewDeploymentKey("test2")
	assert.NoError(t, state.ApplyEvent(&schema.DeploymentCreatedEvent{
		Key: deploymentKey,
		Schema: &schema.Module{
			Name: "test2",
			Runtime: &schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: deploymentKey},
			},
		},
	}))
	assert.NoError(t, state.ApplyEvent(&schema.DeploymentActivatedEvent{
		Key:         deploymentKey,
		MinReplicas: 1,
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

	unmarshalledState := NewSchemaState()
	if err := unmarshalledState.Unmarshal(bytes); err != nil {
		t.Fatalf("failed to unmarshal schema state: %v", err)
	}
	assert.Equal(t, state, unmarshalledState)
}
