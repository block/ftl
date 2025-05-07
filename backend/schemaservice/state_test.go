package schemaservice

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

func TestMarshalling(t *testing.T) {
	t.Run("test roundtrip of single module schema state", func(t *testing.T) {
		k := key.NewDeploymentKey("test", "test")
		state := SchemaState{
			state: &schema.SchemaState{
				Modules: []*schema.Module{{
					Name: "test",
					Runtime: &schema.ModuleRuntime{
						Deployment: &schema.ModuleRuntimeDeployment{
							DeploymentKey: k,
						},
					},
				}},
			},
		}
		assertRoundTrip(t, state)
	})
	t.Run("test roundtrip of schema state after common events", func(t *testing.T) {
		state := NewSchemaState("")

		deploymentKey := key.NewDeploymentKey("test", "test2")
		changesetKey := key.NewChangesetKey()
		ctx := log.ContextWithNewDefaultLogger(t.Context())
		assert.NoError(t, state.ApplyEvent(ctx, &schema.ChangesetCreatedEvent{
			Changeset: &schema.Changeset{
				Key: changesetKey,
				RealmChanges: []*schema.RealmChange{{
					Modules: []*schema.Module{{
						Name: "test2",
						Runtime: &schema.ModuleRuntime{
							Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: deploymentKey},
						},
					},
					},
				}},
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
	})
}

func TestDeploymentEvents(t *testing.T) {
	t.Run("no deployment events on non exiting module", func(t *testing.T) {
		state := NewSchemaState("")
		assert.Equal(t, 0, len(state.DeploymentEvents("foo")))
	})
	t.Run("deployment events on existing module", func(t *testing.T) {
		state := NewSchemaState("")
		state.state.DeploymentEvents = append(state.state.DeploymentEvents, &schema.DeploymentRuntimeEvent{
			Payload: &schema.RuntimeElement{
				Deployment: key.NewDeploymentKey("realm", "foo"),
				Element: &schema.ModuleRuntimeRunner{
					Endpoint: "http://localhost:8080",
				},
			},
		})
		assert.Equal(t, 1, len(state.DeploymentEvents("foo")))
	})
	t.Run("clear deployment events", func(t *testing.T) {
		state := NewSchemaState("")
		state.state.DeploymentEvents = append(state.state.DeploymentEvents, &schema.DeploymentRuntimeEvent{
			Payload: &schema.RuntimeElement{
				Deployment: key.NewDeploymentKey("realm", "foo"),
				Element: &schema.ModuleRuntimeRunner{
					Endpoint: "http://localhost:8080",
				},
			},
		})
		state.clearDeploymentEvents("foo")
		assert.Equal(t, 0, len(state.DeploymentEvents("foo")))
	})
}

func assertRoundTrip(t *testing.T, state SchemaState) {
	t.Helper()
	bytes, err := state.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal schema state: %v", err)
	}

	unmarshalledState := NewSchemaState("")
	if err := unmarshalledState.Unmarshal(bytes); err != nil {
		t.Fatalf("failed to unmarshal schema state: %v", err)
	}
	assert.Equal(t, state, unmarshalledState)
}
