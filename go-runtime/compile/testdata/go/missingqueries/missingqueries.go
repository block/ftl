package missingqueries

import (
	"context"
	"fmt"
	"time"

	"ftl/missingqueries/db"
)

type Price struct {
	Code     string
	Price    float64
	Time     time.Time
	Currency string
}

//ftl:verb export
func SavePrice(ctx context.Context, price Price, client db.InsertPriceClient) error {
	return client(ctx, db.InsertPriceQuery{
		Code:     price.Code,
		Price:    fmt.Sprintf("%.2f", price.Price),
		Time:     price.Time,
		Currency: "USD",
	})
}
