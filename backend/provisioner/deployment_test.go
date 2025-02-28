package provisioner_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	proto "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	"github.com/block/ftl/backend/provisioner"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

// MockProvisioner is a mock implementation of the Provisioner interface
type MockProvisioner struct {
	calls int
}

var _ provisioner.Plugin = (*MockProvisioner)(nil)

func (m *MockProvisioner) Provision(ctx context.Context, req *proto.ProvisionRequest) ([]*schemapb.RuntimeElement, error) {
	m.calls++
	return []*schemapb.RuntimeElement{}, nil
}

func TestDeployment_Progress(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Run("no tasks, no errors", func(t *testing.T) {
		deployment := &provisioner.Deployment{}
		err := deployment.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("runs all provisioners matching the resource type", func(t *testing.T) {
		mock1 := &MockProvisioner{}
		mock2 := &MockProvisioner{}
		mock3 := &MockProvisioner{}

		registry := provisioner.ProvisionerRegistry{}
		registry.Register("mock1", mock1, schema.ResourceTypePostgres)
		registry.Register("mock2", mock2, schema.ResourceTypeMysql)
		registry.Register("mock3", mock3, schema.ResourceTypeSubscription)

		dpl := registry.CreateDeployment(ctx, key.NewChangesetKey(), &schema.Module{
			Name: "testModule",
			Runtime: &schema.ModuleRuntime{
				Deployment: &schema.ModuleRuntimeDeployment{DeploymentKey: key.NewDeploymentKey("test-module")},
			},
			Decls: []schema.Decl{
				&schema.Database{Name: "a", Type: "mysql"},
				&schema.Database{Name: "b", Type: "postgres"},
			},
		}, nil, nil)

		err := dpl.Run(ctx)
		assert.NoError(t, err)

		assert.Equal(t, 1, mock1.calls)
		assert.Equal(t, 1, mock2.calls)
		assert.Equal(t, 0, mock3.calls)
	})
}
