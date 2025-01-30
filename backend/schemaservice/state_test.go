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
