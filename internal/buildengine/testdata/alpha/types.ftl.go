// Code generated by FTL. DO NOT EDIT.
package alpha

import (
	"context"
	ftlother "ftl/other"
	"github.com/block/ftl/common/reflection"
	lib "github.com/block/ftl/go-runtime/schema/testdata"
	"github.com/block/ftl/go-runtime/server"
)

type EchoClient func(context.Context, EchoRequest) (EchoResponse, error)

func init() {
	reflection.Register(
		reflection.ExternalType(*new(lib.AnotherNonFTLType)),

		reflection.ProvideResourcesForVerb(
			Echo,
			server.VerbClient[ftlother.EchoClient, ftlother.EchoRequest, ftlother.EchoResponse](),
		),
	)
}
