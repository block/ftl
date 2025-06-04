package admin

import (
	"context"
	"net"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"

	adminpb "github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/admin/v1/adminpbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
)

// EnvironmentClient standardizes an common interface between the Service as accessed via gRPC
// and a purely-local variant that doesn't require a running controller to access.
type EnvironmentClient interface {
	// List configuration.
	ConfigList(ctx context.Context, req *connect.Request[adminpb.ConfigListRequest]) (*connect.Response[adminpb.ConfigListResponse], error)

	// Get a config value.
	ConfigGet(ctx context.Context, req *connect.Request[adminpb.ConfigGetRequest]) (*connect.Response[adminpb.ConfigGetResponse], error)

	// Set a config value.
	ConfigSet(ctx context.Context, req *connect.Request[adminpb.ConfigSetRequest]) (*connect.Response[adminpb.ConfigSetResponse], error)

	// Unset a config value.
	ConfigUnset(ctx context.Context, req *connect.Request[adminpb.ConfigUnsetRequest]) (*connect.Response[adminpb.ConfigUnsetResponse], error)

	// List secrets.
	SecretsList(ctx context.Context, req *connect.Request[adminpb.SecretsListRequest]) (*connect.Response[adminpb.SecretsListResponse], error)

	// Get a secret.
	SecretGet(ctx context.Context, req *connect.Request[adminpb.SecretGetRequest]) (*connect.Response[adminpb.SecretGetResponse], error)

	// Set a secret.
	SecretSet(ctx context.Context, req *connect.Request[adminpb.SecretSetRequest]) (*connect.Response[adminpb.SecretSetResponse], error)

	// Unset a secret.
	SecretUnset(ctx context.Context, req *connect.Request[adminpb.SecretUnsetRequest]) (*connect.Response[adminpb.SecretUnsetResponse], error)

	// MapConfigsForModule combines all configuration values visible to the module.
	// Local values take precedence.
	MapConfigsForModule(ctx context.Context, req *connect.Request[adminpb.MapConfigsForModuleRequest]) (*connect.Response[adminpb.MapConfigsForModuleResponse], error)

	// MapSecretsForModule combines all secrets visible to the module.
	// Local values take precedence.
	MapSecretsForModule(ctx context.Context, req *connect.Request[adminpb.MapSecretsForModuleRequest]) (*connect.Response[adminpb.MapSecretsForModuleResponse], error)
}

// ShouldUseLocalClient returns whether a local admin client should be used based on the admin service client and the endpoint.
//
// If the service is not present AND endpoint is local, then a local client should be used
// so that the user does not need to spin up a cluster just to run the `ftl config/secret` commands.
//
// If true is returned, use NewLocalClient() to create a local client after setting up config and secret managers for the context.
func ShouldUseLocalClient(ctx context.Context, adminClient adminpbconnect.AdminServiceClient, endpoint *url.URL) (bool, error) {
	isLocal, err := isEndpointLocal(endpoint)
	if err != nil {
		return false, errors.WithStack(err)
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Millisecond*500)
	defer cancel()
	_, err = adminClient.Ping(timeoutCtx, connect.NewRequest(&ftlv1.PingRequest{}))
	if isConnectUnavailableError(err) && isLocal {
		return true, nil
	}
	return false, nil
}

func isConnectUnavailableError(err error) bool {
	var connectErr *connect.Error
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return true
	} else if errors.As(err, &connectErr) {
		return connectErr.Code() == connect.CodeUnavailable
	}
	return false
}

func isEndpointLocal(endpoint *url.URL) (bool, error) {
	h := endpoint.Hostname()
	ips, err := net.LookupIP(h)
	if err != nil {
		return false, errors.Wrap(err, "failed to look up own IP")
	}
	for _, netip := range ips {
		if netip.IsLoopback() {
			return true, nil
		}
	}
	return false, nil
}
