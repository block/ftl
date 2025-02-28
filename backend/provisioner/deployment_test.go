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
type MockProvisioner struct{}

var _ provisioner.Plugin = (*MockProvisioner)(nil)

func (m *MockProvisioner) Provision(ctx context.Context, req *proto.ProvisionRequest) (chan *proto.StatusResponse, error) {
	statusCh := make(chan *proto.StatusResponse, 64)

	statusCh <- &proto.StatusResponse{
		Status: &proto.StatusResponse_Running{},
	}
	statusCh <- &proto.StatusResponse{
		Status: &proto.StatusResponse_Running{},
	}
	statusCh <- &proto.StatusResponse{
		Status: &proto.StatusResponse_Success{Success: &proto.StatusResponse_ProvisioningSuccess{
			Outputs: []*schemapb.RuntimeElement{},
		}},
	}

	close(statusCh)

	return statusCh, nil
}

func TestDeployment_Progress(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	t.Run("no tasks", func(t *testing.T) {
		deployment := &provisioner.Deployment{}
		progress, err := deployment.Progress(ctx)
		assert.NoError(t, err)
		assert.False(t, progress)
	})

	t.Run("progresses each provisioner in order", func(t *testing.T) {
		mock := &MockProvisioner{}

		registry := provisioner.ProvisionerRegistry{}
		registry.Register("mock", mock, schema.ResourceTypePostgres)
		registry.Register("mock", mock, schema.ResourceTypeMysql)

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
		assert.Equal(t, 2, len(dpl.State().Pending))

		_, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(dpl.State().Pending))
		assert.NotEqual(t, 0, len(dpl.State().Done))

		_, err = dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))

		running, err := dpl.Progress(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(dpl.State().Done))
		assert.False(t, running)
	})
}
