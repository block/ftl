package mysql

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type HelloRequest struct {
	Name ftl.Option[string] `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

//ftl:verb export
func Hello(ctx context.Context, req HelloRequest) (HelloResponse, error) {
	return HelloResponse{Message: fmt.Sprintf("Hello, %s!", req.Name.Default("anonymous"))}, nil
}
