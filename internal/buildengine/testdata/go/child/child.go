package child

import (
	"context"
	"ftl/parent"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type HelloRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Hello(ctx context.Context, req parent.HelloRequest, client parent.Verb1Client) (parent.HelloResponse, error) {
	return client(ctx, req)
}
