// Code generated by FTL. DO NOT EDIT.
package shipping

import (
     "github.com/TBD54566975/ftl/go-runtime/ftl"
    "context"
    "github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
    ftlbuiltin "ftl/builtin"
    ftlcurrency "ftl/currency"
)

type GetQuoteClient func(context.Context, ftlbuiltin.HttpRequest[ShippingRequest, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[ftlcurrency.Money, ftl.Unit], error)

type ShipOrderClient func(context.Context, ftlbuiltin.HttpRequest[ShippingRequest, ftl.Unit, ftl.Unit]) (ftlbuiltin.HttpResponse[ShipOrderResponse, ftl.Unit], error)

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
            GetQuote,
		),
		reflection.ProvideResourcesForVerb(
            ShipOrder,
		),
	)
}