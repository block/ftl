// Code generated by FTL. DO NOT EDIT.
package verbtypes

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/server"
)

type CalleeVerbClient func(context.Context, Request) (Response, error)

type CallerVerbClient func(context.Context, Request) (Response, error)

type EmptyClient func(context.Context) error

type SinkClient func(context.Context, Request) error

type SourceClient func(context.Context) (Response, error)

type VerbClient func(context.Context, Request) (Response, error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			CalleeVerb,
		),
		reflection.ProvideResourcesForVerb(
			CallerVerb,
			server.VerbClient[CalleeVerbClient, Request, Response](),
		),
		reflection.ProvideResourcesForVerb(
			Empty,
		),
		reflection.ProvideResourcesForVerb(
			Sink,
		),
		reflection.ProvideResourcesForVerb(
			Source,
		),
		reflection.ProvideResourcesForVerb(
			Verb,
		),
	)
}
