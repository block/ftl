// Code generated by FTL. DO NOT EDIT.
package typeregistry

import (
	"context"
	ftlbuiltin "ftl/builtin"
	ftlsubpackage "ftl/typeregistry/subpackage"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type EchoClient func(context.Context, ftlbuiltin.HttpRequest[EchoRequest, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[EchoResponse, string], error)

func init() {
	reflection.Register(
		reflection.SumType[ftlsubpackage.StringsTypeEnum](
			*new(ftlsubpackage.List),
			*new(ftlsubpackage.Object),
			*new(ftlsubpackage.Single),
		),
		reflection.ProvideResourcesForVerb(
			Echo,
		),
	)
}
