package proxy

import (
	"context"

	"ftl/echo"
)

// Proxy returns a greeting
//
//ftl:verb export
func Proxy(ctx context.Context, req string, echo echo.EchoClient) (string, error) {
	return echo(ctx, req)
}

//ftl:verb export
func ProxyCallerIdentity(ctx context.Context, client echo.CallerIdentityClient) (string, error) {
	return client(ctx)
}