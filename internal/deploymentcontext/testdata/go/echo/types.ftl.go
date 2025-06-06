// Code generated by FTL. DO NOT EDIT.
package echo

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/server"
)

type HelloClient func(context.Context, HelloRequest) (HelloResponse, error)

func init() {
	reflection.Register(

		reflection.ProvideResourcesForVerb(
			Hello,
			server.Config[string]("echo", "greeting"),
			server.Secret[string]("echo", "apiKey"),
		),
	)
}
