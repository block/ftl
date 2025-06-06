package provisioner

import (
	"context"

	"connectrpc.com/connect"

	provisioner "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1/provisionerpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
)

// NoopProvisioner is a provisioner that does nothing
type NoopProvisioner struct{}

var _ provisionerconnect.ProvisionerPluginServiceClient = (*NoopProvisioner)(nil)

func (d *NoopProvisioner) Ping(context.Context, *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return &connect.Response[ftlv1.PingResponse]{}, nil
}

func (d *NoopProvisioner) Provision(ctx context.Context, req *connect.Request[provisioner.ProvisionRequest]) (*connect.Response[provisioner.ProvisionResponse], error) {
	return connect.NewResponse(&provisioner.ProvisionResponse{
		Tasks: []*provisioner.ProvisioningStatus{{
			Status: &provisioner.ProvisioningStatus_Running{
				Running: &provisioner.ProvisioningStatus_ProvisioningRunning{
					ProvisioningToken: "token",
				},
			},
		}},
	}), nil
}

func (d *NoopProvisioner) Status(ctx context.Context, req *connect.Request[provisioner.StatusRequest]) (*connect.Response[provisioner.StatusResponse], error) {
	return connect.NewResponse(&provisioner.StatusResponse{
		Status: &provisioner.ProvisioningStatus{
			Status: &provisioner.ProvisioningStatus_Success{
				Success: &provisioner.ProvisioningStatus_ProvisioningSuccess{},
			},
		},
	}), nil
}
