package mysql

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type MyDbConfig struct {
	ftl.DefaultMySQLDatabaseConfig
}

func (MyDbConfig) Name() string { return "testdb" }

type InsertRequest struct {
	Data string
}

type InsertResponse struct{}

//ftl:verb
func Insert(ctx context.Context, req InsertRequest, createRequest CreateRequestClient) (InsertResponse, error) {
	err := createRequest(ctx, CreateRequestQuery{Data: req.Data})
	if err != nil {
		return InsertResponse{}, err
	}

	return InsertResponse{}, nil
}

//ftl:verb
func Query(ctx context.Context, getRequestData GetRequestDataClient) (map[string]string, error) {
	result, err := getRequestData(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]string{"data": result[0].Data}, nil
}
