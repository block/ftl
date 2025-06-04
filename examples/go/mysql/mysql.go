package mysql

import (
	"context"
	"ftl/mysql/db"

	"github.com/alecthomas/errors" // Import the FTL SDK.

	"github.com/block/ftl/go-runtime/ftl"
)

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb export
func Insert(ctx context.Context, req InsertRequest, insert db.CreateRequestClient) (InsertResponse, error) {
	err := insert(ctx, ftl.Some(req.Data))
	if err != nil {
		return InsertResponse{}, errors.WithStack(err)
	}

	return InsertResponse{}, nil
}

//ftl:verb export
func Query(ctx context.Context, query db.GetRequestDataClient) ([]string, error) {
	rows, err := query(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var items []string
	for _, row := range rows {
		if d, ok := row.Get(); ok {
			items = append(items, d)
		}
	}
	return items, nil
}

//ftl:fixture
func Fixture(ctx context.Context, insert db.CreateRequestClient, query db.GetRequestDataClient) error {
	rows, err := query(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(rows) > 0 {
		return nil
	}
	return errors.WithStack(errors.Join(
		insert(ctx, ftl.Some("foo")),
		insert(ctx, ftl.Some("bar")),
	))
}
