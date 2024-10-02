// Code generated by FTL. DO NOT EDIT.
package pubsub

import (
	"context"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

type BroadcastClient func(context.Context) error

type PayinClient func(context.Context) error

type ProcessBroadcastClient func(context.Context, PayinEvent) error

type ProcessPayinClient func(context.Context, PayinEvent) error

func init() {
	reflection.Register(
		reflection.ProvideResourcesForVerb(
			Broadcast,
		),
		reflection.ProvideResourcesForVerb(
			Payin,
		),
		reflection.ProvideResourcesForVerb(
			ProcessBroadcast,
		),
		reflection.ProvideResourcesForVerb(
			ProcessPayin,
		),
	)
}
