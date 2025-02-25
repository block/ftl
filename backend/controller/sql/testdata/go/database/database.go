package database

import (
	"context"
)

type InsertRequest struct {
	Data string
	Id   int
}

type InsertResponse struct{}

//ftl:verb
func Insert(ctx context.Context, req InsertRequest, db TestdbHandle) (InsertResponse, error) {
	err := persistRequest(ctx, req, db)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

func persistRequest(ctx context.Context, req InsertRequest, db TestdbHandle) error {
	_, err := db.Get(ctx).Exec("INSERT INTO requests (id,data) VALUES ($1, $2);", req.Id, req.Data)
	if err != nil {
		return err
	}
	return nil
}
