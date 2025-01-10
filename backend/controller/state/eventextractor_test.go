package state

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/tuple"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/model"
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
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfas": {
						Module:    "test",
						Key:       deploymentKey(t, "dpl-test-sjkfislfjslfas"),
						CreatedAt: now,
						Schema:    &schema.Module{Name: "test"},
						Language:  "go",
					},
				},
			},
			want: []SchemaEvent{
				&DeploymentCreatedEvent{
					Module:    "test",
					Key:       deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					CreatedAt: now,
					Schema:    &schema.Module{Name: "test"},
					Language:  "go",
				},
			},
		},
		{
			name: "schema update creates schema updated event",
			previous: SchemaState{
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfas": {
						Module:    "test",
						Key:       deploymentKey(t, "dpl-test-sjkfislfjslfas"),
						CreatedAt: now,
						Schema:    &schema.Module{Name: "test"},
						Language:  "go",
					},
				},
			},
			current: SchemaState{
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfas": {
						Module: "test",
						Key:    deploymentKey(t, "dpl-test-sjkfislfjslfas"),
						Schema: &schema.Module{Name: "test", Metadata: []schema.Metadata{&schema.MetadataArtefact{}}},
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
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfas": {
						Module: "test",
						Key:    deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					},
				},
				activeDeployments: map[string]bool{
					"dpl-test-sjkfislfjslfas": true,
				},
			},
			current: SchemaState{
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfas": {
						Module: "test",
						Key:    deploymentKey(t, "dpl-test-sjkfislfjslfas"),
					},
				},
				activeDeployments: map[string]bool{},
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
				deployments: map[string]*Deployment{
					"dpl-test-sjkfislfjslfaa": {
						Module: "test",
						Key:    deploymentKey(t, "dpl-test-sjkfislfjslfaa"),
					},
				},
				activeDeployments: map[string]bool{
					"dpl-test-sjkfislfjslfaa": true,
				},
			},
			current: SchemaState{
				deployments: map[string]*Deployment{},
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
			got := EventExtractor(tuple.PairOf(tt.previous, tt.current))
			assert.Equal(t, tt.want, got)
		})
	}
}

func deploymentKey(t *testing.T, name string) model.DeploymentKey {
	t.Helper()
	key, err := model.ParseDeploymentKey(name)
	if err != nil {
		t.Fatalf("failed to parse deployment key: %v", err)
	}
	return key
}
