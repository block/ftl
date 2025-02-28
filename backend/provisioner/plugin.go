package provisioner

import (
	"context"

	"connectrpc.com/connect"
	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
)

// Plugin
type Plugin interface {
	Provision(ctx context.Context, req *provisioner.ProvisionRequest) (chan *provisioner.StatusResponse, error)
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
func (p *PluginGrpcClient) Provision(ctx context.Context, req *provisioner.ProvisionRequest) (chan *provisioner.StatusResponse, error) {
	resp, err := p.client.Provision(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}

	// call status endpoint in a loop until the status is not pending
	statusCh := make(chan *provisioner.StatusResponse)
	go func() {
		for {
			status, err := p.client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{ProvisioningToken: resp.Msg.ProvisioningToken}))
			if err != nil {
				return
			}
			statusCh <- status.Msg

			// terminate the channel when the task is done
			if status.Msg.GetRunning() == nil {
				close(statusCh)
				return
			}
		}
	}()

	return statusCh, nil
}
