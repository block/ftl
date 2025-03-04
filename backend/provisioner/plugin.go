package provisioner

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
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
		return nil, fmt.Errorf("error provisioning: %w", err)
	}

	// call status endpoint in a loop until the status is not pending
	statusCh := make(chan *provisioner.StatusResponse)
	backoff := backoff.Backoff{
		Min: 50 * time.Millisecond,
		Max: 30 * time.Second,
	}
	for {
		status, err := p.client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
			ProvisioningToken: resp.Msg.ProvisioningToken,
			DesiredModule:     req.DesiredModule,
		}))
		if err != nil {
			statusCh <- &provisioner.StatusResponse{
				Status: &provisioner.StatusResponse_Failed{
					Failed: &provisioner.StatusResponse_ProvisioningFailed{
						ErrorMessage: err.Error(),
					},
				},
			}
		}

		switch status.Msg.GetStatus().(type) {
		case *provisioner.StatusResponse_Running:
			time.Sleep(backoff.Duration())
		case *provisioner.StatusResponse_Success:
			return status.Msg.GetSuccess().Outputs, nil
		case *provisioner.StatusResponse_Failed:
			return nil, fmt.Errorf("provisioning failed: %s", status.Msg.GetFailed().ErrorMessage)
		}
	}
}
