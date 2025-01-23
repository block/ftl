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

	tests := []struct {
		name     string
		previous SchemaState
		current  SchemaState
		want     []SchemaEvent
	}{
		{
			name:     "new deployment creates deployment event",
			previous: SchemaState{},
			current: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): {
						Name: "test",
						Runtime: &schema.ModuleRuntime{
							Base: schema.ModuleRuntimeBase{Language: "go"},
							Deployment: &schema.ModuleRuntimeDeployment{
								CreatedAt: now,
							},
						},
					},
				},
			},
			want: []SchemaEvent{
				&DeploymentCreatedEvent{
					Key: deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					Schema: &schema.Module{Name: "test", Runtime: &schema.ModuleRuntime{
						Base: schema.ModuleRuntimeBase{Language: "go"},
						Deployment: &schema.ModuleRuntimeDeployment{
							CreatedAt: now,
						},
					}},
				},
			},
		},
		{
			name: "schema update creates schema updated event",
			previous: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): {
						Name: "test",
					},
				},
			},
			current: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): {
						Name:     "test",
						Metadata: []schema.Metadata{&schema.MetadataArtefact{}},
					},
				},
			},
			want: []SchemaEvent{
				&DeploymentSchemaUpdatedEvent{
					Key:    deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					Schema: &schema.Module{Name: "test", Metadata: []schema.Metadata{&schema.MetadataArtefact{}}},
				},
			},
		},
		{
			name: "deactivated deployment creates deactivation event",
			previous: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): {
						Name: "test",
					},
				},
				activeDeployments: map[key.Deployment]bool{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): true,
				},
			},
			current: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfas"): {
						Name: "test",
					},
				},
				activeDeployments: map[key.Deployment]bool{},
			},
			want: []SchemaEvent{
				&DeploymentDeactivatedEvent{
					Key:           deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					ModuleRemoved: false,
				},
			},
		}, {
			name: "removing an active deployment creates module removed event",
			previous: SchemaState{
				deployments: map[key.Deployment]*schema.Module{
					deploymentKey(t, "dpl-test-sjkfislfjslfaa"): {
						Name: "test",
					},
				},
				activeDeployments: map[key.Deployment]bool{
					deploymentKey(t, "dpl-test-sjkfislfjslfaa"): true,
				},
			},
			current: SchemaState{
				deployments: map[key.Deployment]*schema.Module{},
			},
			want: []SchemaEvent{
				&DeploymentDeactivatedEvent{
					Key:           deploymentKey(t, "dpl-test-sjkfislfjslfaa"),
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
