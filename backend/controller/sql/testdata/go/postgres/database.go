package postgres

import (
	"context"
	"errors"

	"github.com/block/ftl/go-runtime/ftl"
)

type InsertRequest struct {
	Data string
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

type TransactionRequest struct {
	Items []string
}

type TransactionResponse struct {
	Count int
}

//ftl:transaction
func TransactionInsert(ctx context.Context, req TransactionRequest, createRequest CreateRequestClient, getRequestData GetRequestDataClient) (TransactionResponse, error) {
	for _, item := range req.Items {
		err := createRequest(ctx, ftl.Some(item))
		if err != nil {
			return TransactionResponse{}, err
		}
	}
	result, err := getRequestData(ctx)
	if err != nil {
		return TransactionResponse{}, err
	}
	return TransactionResponse{Count: len(result)}, nil
}

//ftl:transaction
func TransactionRollback(ctx context.Context, req TransactionRequest, createRequest CreateRequestClient) (TransactionResponse, error) {
	if len(req.Items) > 0 {
		err := createRequest(ctx, ftl.Some(req.Items[0]))
		if err != nil {
			return TransactionResponse{}, err
		}
	}
	return TransactionResponse{}, errors.New("deliberate error to test rollback")
}

func persistRequest(ctx context.Context, req InsertRequest, db TestdbHandle) error {
	_, err := db.Get(ctx).Exec("INSERT INTO requests (data) VALUES ($1);", req.Data)
	if err != nil {
		return err
	}
	return nil
}
