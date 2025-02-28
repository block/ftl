package provisioner

import (
	"context"

	"connectrpc.com/connect"
	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
)

// Plugin
type Plugin interface {
	Provision(ctx context.Context, req *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error)
	Status(ctx context.Context, req *provisioner.StatusRequest) (*provisioner.StatusResponse, error)
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
func (p *PluginGrpcClient) Provision(ctx context.Context, req *provisioner.ProvisionRequest) (*provisioner.ProvisionResponse, error) {
	resp, err := p.client.Provision(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// Status implements the Plugin interface
func (p *PluginGrpcClient) Status(ctx context.Context, req *provisioner.StatusRequest) (*provisioner.StatusResponse, error) {
	resp, err := p.client.Status(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}
