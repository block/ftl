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

	// TODO: lots of tests once we have new runtime events
	empty := ""
	newKey, err := key.ParseDeploymentKey("dpl-test-test-sjkfislfjslfae")
	assert.NoError(t, err)
	tests := []struct {
		name     string
		previous SchemaState
		current  SchemaState
		want     []*ftlv1.PullSchemaResponse
	}{
		{
			name:     "new deployment creates deployment event",
			previous: SchemaState{state: &schema.SchemaState{}},
			current: SchemaState{
				state: &schema.SchemaState{
					Modules: []*schema.Module{
						{
							Name: "test",
							Runtime: &schema.ModuleRuntime{
								Base: schema.ModuleRuntimeBase{Language: "go", CreateTime: now},
								Deployment: &schema.ModuleRuntimeDeployment{
									CreatedAt:     now,
									DeploymentKey: newKey,
								},
							},
						},
					},
					Realms: []*schema.RealmState{
						{Name: "test"},
					},
				},
			},
			want: []*ftlv1.PullSchemaResponse{
				{
					Event: &schemapb.Notification{Value: &schemapb.Notification_FullSchemaNotification{
						FullSchemaNotification: &schemapb.FullSchemaNotification{
							Schema: &schemapb.Schema{
								Pos: &schemapb.Position{},
								Realms: []*schemapb.Realm{{
									Pos:  &schemapb.Position{},
									Name: "test",
									Modules: []*schemapb.Module{{Name: "test",
										Pos: &schemapb.Position{},
										Runtime: &schemapb.ModuleRuntime{
											Base: &schemapb.ModuleRuntimeBase{Language: "go", Os: &empty, Arch: &empty, Image: &empty, CreateTime: timestamppb.New(now)},
											Deployment: &schemapb.ModuleRuntimeDeployment{
												CreatedAt:     timestamppb.New(now),
												DeploymentKey: newKey.String(),
											},
										},
									}},
								}},
							},
						}},
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
