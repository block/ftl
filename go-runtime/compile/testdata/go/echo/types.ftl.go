// Code generated by FTL. DO NOT EDIT.
package echo

import (
	"context"
	ftltime "ftl/time"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type EchoClient func(context.Context, EchoRequest) (EchoResponse, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Echo,
			server.VerbClient[ftltime.TimeClient, ftltime.TimeRequest, ftltime.TimeResponse](),
			server.Config[string]("echo", "defaultName"),
		),
	)
}
