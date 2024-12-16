// Code generated by FTL. DO NOT EDIT.
package two

import (
	"context"
	ftlbuiltin "ftl/builtin"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/common/reflection"
	lib "github.com/block/ftl/go-runtime/schema/testdata"
	"github.com/block/ftl/go-runtime/server"
	"github.com/jpillora/backoff"
)

type CallsTwoClient func(context.Context, Payload[string]) (Payload[string], error)

type TwoClient func(context.Context, Payload[string]) (Payload[string], error)

type CallsTwoAndThreeClient func(context.Context, Payload[string]) (Payload[string], error)

type ThreeClient func(context.Context, Payload[string]) (Payload[string], error)

type IngressClient func(context.Context, ftlbuiltin.HttpRequest[PostRequest, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[PostResponse, string], error)

type ReturnsUserClient func(context.Context) (UserResponse, error)

func init() {
	reflection.Register(
		reflection.SumType[TypeEnum](
			*new(Exported),
			*new(List),
			*new(Scalar),
			*new(WithoutDirective),
		),
		reflection.ExternalType(*new(backoff.Backoff)),
		reflection.ExternalType(*new(lib.NonFTLType)),
		reflection.Database[FooConfig]("foo", server.InitPostgres),
		reflection.ProvideResourcesForVerb(
			CallsTwo,
			server.VerbClient[TwoClient, Payload[string], Payload[string]](),
		),
		reflection.ProvideResourcesForVerb(
			Two,
			server.DatabaseHandle[FooConfig]("postgres"),
		),
		reflection.ProvideResourcesForVerb(
			CallsTwoAndThree,
			server.VerbClient[TwoClient, Payload[string], Payload[string]](),
			server.VerbClient[ThreeClient, Payload[string], Payload[string]](),
		),
		reflection.ProvideResourcesForVerb(
			Three,
		),
		reflection.ProvideResourcesForVerb(
			Ingress,
		),
		reflection.ProvideResourcesForVerb(
			ReturnsUser,
		),
	)
}
