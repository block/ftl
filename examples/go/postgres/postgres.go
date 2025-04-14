package postgres

import (
	"context"

	errors "github.com/alecthomas/errors"
)

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb export
func Insert(ctx context.Context, req InsertRequest, db TestdbHandle) (InsertResponse, error) {
	err := persistRequest(ctx, req, db)
	if err != nil {
		return InsertResponse{}, errors.WithStack(err)
	}

	return InsertResponse{}, nil
}

//ftl:verb export
func Query(ctx context.Context, db TestdbHandle) ([]string, error) {
	rows, err := db.Get(ctx).QueryContext(ctx, "SELECT data FROM requests")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var i string
		if err := rows.Scan(
			&i,
		); err != nil {
			return nil, errors.WithStack(err)
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return items, nil
}

func persistRequest(ctx context.Context, req InsertRequest, db TestdbHandle) error {
	_, err := db.Get(ctx).Exec("INSERT INTO requests (data) VALUES ($1)", req.Data)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
