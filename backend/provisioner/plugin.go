package provisioner

import (
	"context"
	"strings"
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
	tasks := resp.Msg.Tasks
	var results []*schemapb.RuntimeElement
	var failures []string

	// call status endpoint in a loop until the status is not pending
	backoff := backoff.Backoff{
		Min: 50 * time.Millisecond,
		Max: 30 * time.Second,
	}
	for {
		var tokens []string
		for _, task := range tasks {
			if task.GetRunning() != nil {
				tokens = append(tokens, task.GetRunning().GetProvisioningToken())
			} else if task.GetSuccess() != nil {
				results = append(results, task.GetSuccess().GetOutputs()...)
			} else if task.GetFailed() != nil {
				failures = append(failures, task.GetFailed().GetErrorMessage())
			}
		}

		if len(tokens) == 0 {
			if len(failures) > 0 {
				return nil, errors.Errorf("provisioning failed: %s", strings.Join(failures, ", "))
			}
			return results, nil
		}
		time.Sleep(backoff.Duration())

		tasks = nil
		for _, token := range tokens {
			status, err := p.client.Status(ctx, connect.NewRequest(&provisioner.StatusRequest{
				ProvisioningToken: token,
				DesiredModule:     req.DesiredModule,
			}))
			if err != nil {
				return nil, errors.Wrap(err, "provisioner status check failed")
			}
			tasks = append(tasks, status.Msg.Status)
		}
	}
}
