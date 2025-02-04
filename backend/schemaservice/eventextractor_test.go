package schemaservice

import (
	"slices"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/tuple"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
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
		want     []*ftlv1.PullSchemaResponse
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
			want: []*ftlv1.PullSchemaResponse{
				{
					Event: &ftlv1.PullSchemaResponse_DeploymentCreated_{
						DeploymentCreated: &ftlv1.PullSchemaResponse_DeploymentCreated{
							Schema: &schemapb.Module{Name: "test", Runtime: &schemapb.ModuleRuntime{
								Base: &schemapb.ModuleRuntimeBase{Language: "go"},
								Deployment: &schemapb.ModuleRuntimeDeployment{
									CreatedAt:     timestamppb.New(now),
									DeploymentKey: newKey.String(),
								},
							}},
						},
					},
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
						Runtime: &schema.ModuleRuntime{Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: oldKey, Endpoint: "http://localhost:8080"}},
						Name:    "test",
					},
				},
			},
			want: []*ftlv1.PullSchemaResponse{
				{
					Event: &ftlv1.PullSchemaResponse_DeploymentUpdated_{
						DeploymentUpdated: &ftlv1.PullSchemaResponse_DeploymentUpdated{
							Schema: &schemapb.Module{
								Name: "test",
								Runtime: &schemapb.ModuleRuntime{
									Base: &schemapb.ModuleRuntimeBase{Language: "go"},
									Deployment: &schemapb.ModuleRuntimeDeployment{
										CreatedAt:     timestamppb.New(now),
										DeploymentKey: newKey.String(),
									},
								}},
						},
					},
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
			want: []*ftlv1.PullSchemaResponse{
				{
					Event: &ftlv1.PullSchemaResponse_DeploymentUpdated_{
						DeploymentUpdated: &ftlv1.PullSchemaResponse_DeploymentUpdated{
							Schema: &schemapb.Module{
								Name: "test",
								Runtime: &schemapb.ModuleRuntime{
									Base: &schemapb.ModuleRuntimeBase{Language: "go"},
									Deployment: &schemapb.ModuleRuntimeDeployment{
										CreatedAt:     timestamppb.New(now),
										DeploymentKey: newKey.String(),
									},
								}},
						},
					},
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
