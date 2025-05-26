package provisioner

import (
	"context"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	"github.com/jpillora/backoff"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1/provisionerpbconnect"
	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

// Plugin that can be used to provision resources
type Plugin interface {
	Provision(ctx context.Context, req *provisioner.ProvisionRequest) ([]*schemapb.RuntimeElement, error)
}

// PluginGrpcClient implements the Plugin interface by wrapping a ProvisionerPluginService client
type PluginGrpcClient struct {
	client provisionerconnect.ProvisionerPluginServiceClient
}

// NewPluginClient creates a new Plugin that wraps a ProvisionerPluginService client
func NewPluginClient(client provisionerconnect.ProvisionerPluginServiceClient) Plugin {
	return &PluginGrpcClient{client: client}
}

// Provision implements the Plugin interface
func (p *PluginGrpcClient) Provision(ctx context.Context, req *provisioner.ProvisionRequest) ([]*schemapb.RuntimeElement, error) {
	resp, err := p.client.Provision(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, errors.Wrap(err, "error provisioning")
	}
	status := resp.Msg.Status
	token := resp.Msg.ProvisioningToken

	// call status endpoint in a loop until the status is not pending
	backoff := backoff.Backoff{
		Min: 50 * time.Millisecond,
		Max: 30 * time.Second,
	}
	for {
		switch s := status.Status.(type) {
		case *provisioner.ProvisioningStatus_Running:
			time.Sleep(backoff.Duration())
		case *provisioner.ProvisioningStatus_Success:
			return s.Success.Outputs, nil
		case *provisioner.ProvisioningStatus_Failed:
			return nil, errors.Errorf("provisioning failed: %s", s.Failed.ErrorMessage)
		}

		statusResp, err := p.client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
			ProvisioningToken: token,
			DesiredModule:     req.DesiredModule,
		}))
		if err != nil {
			return nil, errors.Wrap(err, "provisioner status check faile")
		}
		status = statusResp.Msg.Status
	}
}
