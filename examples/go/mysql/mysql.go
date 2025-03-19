package mysql

import (
	"context"
	"errors"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb export
func Insert(ctx context.Context, req InsertRequest, insert CreateRequestClient) (InsertResponse, error) {
	err := insert(ctx, CreateRequestQuery{Data: ftl.Some(req.Data)})
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

//ftl:verb export
func Query(ctx context.Context, query GetRequestDataClient) ([]string, error) {
	rows, err := query(ctx)
	if err != nil {
		return nil, err
	}
	var items []string
	for _, row := range rows {
		if d, ok := row.Data.Get(); ok {
			items = append(items, d)
		}
	}
	return items, nil
}

//ftl:fixture
func Fixture(ctx context.Context, insert CreateRequestClient, query GetRequestDataClient) error {
	rows, err := query(ctx)
	if err != nil {
		return err
	}
	if len(rows) > 0 {
		return nil
	}
	return errors.Join(
		insert(ctx, CreateRequestQuery{Data: ftl.Some("foo")}),
		insert(ctx, CreateRequestQuery{Data: ftl.Some("bar")}),
	)
}
