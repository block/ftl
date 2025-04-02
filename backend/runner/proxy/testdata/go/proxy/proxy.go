package proxy

import (
	"context"
	"ftl/echo"
	// Import the FTL SDK.
)

//ftl:verb export
func Proxy(ctx context.Context, msg string, echo echo.EchoClient) (string, error) {
	return echo(ctx, msg)
}
