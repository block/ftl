// Code generated by FTL. DO NOT EDIT.
package wrapped

import (
	"context"
	ftltime "ftl/time"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/server"
)

type InnerClient func(context.Context) (WrappedResponse, error)

type OuterClient func(context.Context) (WrappedResponse, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Inner,
			server.VerbClient[ftltime.TimeClient, ftltime.TimeRequest, ftltime.TimeResponse](),
		),
		reflection.ProvideResourcesForVerb(
			Outer,
			server.SourceClient[InnerClient, WrappedResponse](),
		),
	)
}
