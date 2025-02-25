package missingqueries

import (
	"context"
	"fmt"
	"time"
	// Import the FTL SDK.
)

type Price struct {
	Code     string
	Price    float64
	Time     time.Time
	Currency string
}

//ftl:verb export
func SavePrice(ctx context.Context, price Price, client InsertPriceClient) error {
	return client(ctx, InsertPriceQuery{
		Code:     price.Code,
		Price:    fmt.Sprintf("%.2f", price.Price),
		Time:     price.Time,
		Currency: "USD",
	})
}
