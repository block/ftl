package missingqueries

import (
	"context"
	"fmt"
	"time"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type Price struct {
	Code     string
	Price    float64
	Time     time.Time
	Currency string
}

type PriceDatasource struct {
	ftl.DefaultMySQLDatabaseConfig
}

func (PriceDatasource) Name() string { return "prices" }

//ftl:verb export
func SavePrice(ctx context.Context, price Price, client InsertPriceClient) error {
	return client(ctx, InsertPriceQuery{
		Code:     price.Code,
		Price:    fmt.Sprintf("%.2f", price.Price),
		Time:     price.Time,
		Currency: "USD",
	})
}
