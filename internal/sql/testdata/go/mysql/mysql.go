package mysql

import (
	"context"
	"ftl/mysql/db"
)

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb
func Insert(ctx context.Context, req InsertRequest, createRequest db.CreateRequestClient) (InsertResponse, error) {
	err := createRequest(ctx, req.Data)
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

//ftl:verb
func Query(ctx context.Context, getRequestData db.GetRequestDataClient) (map[string]string, error) {
	result, err := getRequestData(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]string{"data": result[0]}, nil
}
