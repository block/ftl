package schemaservice

import (
	"slices"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/tuple"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

func TestEventExtractor(t *testing.T) {
	now := time.Now()

	oldKey, err := key.ParseDeploymentKey("dpl-test-sjkfislfjslfas")
	assert.NoError(t, err)
	newKey, err := key.ParseDeploymentKey("dpl-test-sjkfislfjslfae")
	assert.NoError(t, err)
	tests := []struct {
		name     string
		previous SchemaState
		current  SchemaState
		want     []schema.Event
	}{
		{
			name:     "new deployment creates deployment event",
			previous: SchemaState{},
			current: SchemaState{
				deployments: map[string]*schema.Module{
					"test": {
						Name: "test",
						Runtime: &schema.ModuleRuntime{
							Base: schema.ModuleRuntimeBase{Language: "go"},
							Deployment: &schema.ModuleRuntimeDeployment{
								CreatedAt:     now,
								DeploymentKey: newKey,
							},
						},
					},
				},
			},
			want: []schema.Event{
				&schema.DeploymentCreatedEvent{
					Key: newKey,
					Schema: &schema.Module{Name: "test", Runtime: &schema.ModuleRuntime{
						Base: schema.ModuleRuntimeBase{Language: "go"},
						Deployment: &schema.ModuleRuntimeDeployment{
							CreatedAt:     now,
							DeploymentKey: newKey,
						},
					}},
				},
			},
		},
		{
			name: "schema update creates schema updated event",
			previous: SchemaState{
				deployments: map[string]*schema.Module{
					"test": {
						Name:    "test",
						Runtime: &schema.ModuleRuntime{Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: oldKey}},
					},
				},
			},
			current: SchemaState{
				deployments: map[string]*schema.Module{
					"test": {
						Runtime:  &schema.ModuleRuntime{Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: oldKey}},
						Name:     "test",
						Metadata: []schema.Metadata{&schema.MetadataArtefact{}},
					},
				},
			},
			want: []schema.Event{
				&schema.DeploymentSchemaUpdatedEvent{
					Key: deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					Schema: &schema.Module{Name: "test", Runtime: &schema.ModuleRuntime{Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: oldKey}},
						Metadata: []schema.Metadata{&schema.MetadataArtefact{}}},
				},
			},
		},
		{
			name: "removing an active deployment creates module removed event",
			previous: SchemaState{
				deployments: map[string]*schema.Module{
					"test": {
						Name: "test",
						Runtime: &schema.ModuleRuntime{
							Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: oldKey},
						},
					},
				},
			},
			current: SchemaState{
				deployments: map[string]*schema.Module{},
			},
			want: []schema.Event{
				&schema.DeploymentDeactivatedEvent{
					Key:           oldKey,
					ModuleRemoved: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Collect(EventExtractor(tuple.PairOf(tt.previous, tt.current)))
			assert.Equal(t, tt.want, got)
		})
	}
}

func deploymentKey(t *testing.T, name string) key.Deployment {
	t.Helper()
	key, err := key.ParseDeploymentKey(name)
	if err != nil {
		t.Fatalf("failed to parse deployment key: %v", err)
	}
	return key
}
