package query

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/tuple"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
)

func One[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) (resp Resp, err error) {
	results, err := performQuery[Resp](ctx, dbName, querypb.CommandType_COMMAND_TYPE_ONE, rawSQL, params, colFieldNames)
	if err != nil {
		return resp, err
	}
	if len(results) == 0 {
		return resp, fmt.Errorf("no results found")
	}
	return results[0], nil
}

func Many[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]Resp, error) {
	return performQuery[Resp](ctx, dbName, querypb.CommandType_COMMAND_TYPE_MANY, rawSQL, params, colFieldNames)
}

func Exec[Req any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) error {
	_, err := performQuery[any](ctx, dbName, querypb.CommandType_COMMAND_TYPE_EXEC, rawSQL, params, colFieldNames)
	return err
}

type client struct {
	queryconnect.QueryServiceClient
}

func newClient(dbName string) (*client, error) {
	address := os.Getenv(strings.ToUpper("FTL_QUERY_ADDRESS_" + dbName))
	if address == "" {
		return nil, fmt.Errorf("query address for %s not found", dbName)
	}
	return &client{QueryServiceClient: rpc.Dial(queryconnect.NewQueryServiceClient, address, log.Error)}, nil
}

// performQuery performs a SQL query and returns the results.
//
// note: accepts colFieldNames as []tuple.Pair[string, string] rather than map[string]string to preserve column order
func performQuery[T any](ctx context.Context, dbName string, commandType querypb.CommandType, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]T, error) {
	client, err := newClient(dbName)
	if err != nil {
		return nil, err
	}

	paramsJSON, err := encoding.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parameters: %w", err)
	}

	resultCols := make([]*querypb.ResultColumn, 0, len(colFieldNames))
	for _, pair := range colFieldNames {
		resultCols = append(resultCols, &querypb.ResultColumn{
			SqlName:  pair.A,
			TypeName: pair.B,
		})
	}

	stream, err := client.QueryServiceClient.ExecuteQuery(ctx, connect.NewRequest(&querypb.ExecuteQueryRequest{
		RawSql:         rawSQL,
		CommandType:    commandType,
		ParametersJson: string(paramsJSON),
		ResultColumns:  resultCols,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var results []T
	for stream.Receive() {
		resp := stream.Msg()
		switch r := resp.Result.(type) {
		case *querypb.ExecuteQueryResponse_ExecResult:
			return nil, nil
		case *querypb.ExecuteQueryResponse_RowResults:
			if r.RowResults.JsonRows == "" {
				continue
			}
			var result T
			if err := encoding.Unmarshal([]byte(r.RowResults.JsonRows), &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal row data: %w", err)
			}
			results = append(results, result)
		}
	}
	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("failed to receive query results: %w", err)
	}
	return results, nil
}
