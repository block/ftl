// Code generated by FTL. DO NOT EDIT.
package missingqueries

import (
	"context"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server"
)

type SavePriceClient func(context.Context, Price) error

//ftl:database mysql prices
type PricesConfig struct{}

type PricesHandle = ftl.DatabaseHandle[PricesConfig]

func init() {
	reflection.Register(
		reflection.Database[PricesConfig]("prices", server.InitMySQL),

		reflection.ProvideResourcesForVerb(
			SavePrice,
			server.SinkClient[InsertPriceClient, InsertPriceQuery](),
		),
	)
}
