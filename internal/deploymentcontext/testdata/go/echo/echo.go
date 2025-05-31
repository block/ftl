package echo

import (
	"context"
	"fmt"

	"github.com/block/ftl/go-runtime/ftl"
)

type HelloRequest struct {
	Name string `json:"name"`
}

type HelloResponse struct {
	Message string `json:"message"`
}

type Greeting = ftl.Config[string]
type ApiKey = ftl.Secret[string]

//ftl:verb export
func Hello(ctx context.Context, req HelloRequest, greeting Greeting, apiKey ApiKey) (HelloResponse, error) {
	return HelloResponse{
		Message: fmt.Sprintf("%s, %s!!! (authenticated with %s)",
			greeting.Get(ctx),
			req.Name,
			apiKey.Get(ctx)),
	}, nil
}
