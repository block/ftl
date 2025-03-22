package query

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/tuple"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/go-runtime/server/rpccontext"
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

// performQuery performs a SQL query and returns the results.
//
// note: accepts colFieldNames as []tuple.Pair[string, string] rather than map[string]string to preserve column order
func performQuery[T any](ctx context.Context, dbName string, commandType querypb.CommandType, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]T, error) {
	client := rpccontext.ClientFromContext[queryconnect.QueryServiceClient](ctx)

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

	stream, err := client.ExecuteQuery(ctx, connect.NewRequest(&querypb.ExecuteQueryRequest{
		DatabaseName:   dbName,
		RawSql:         rawSQL,
		CommandType:    commandType,
		ParametersJson: string(paramsJSON),
		ResultColumns:  resultCols,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	responses := []*querypb.ExecuteQueryResponse{}
	if inline, ok := client.(*InlineQueryClient); ok {
		responses = inline.drainResponses()
	} else {
		for stream.Receive() {
			resp := stream.Msg()
			responses = append(responses, resp)
		}
		if err := stream.Err(); err != nil {
			return nil, fmt.Errorf("failed to receive query results: %w", err)
		}
	}

	var results []T
	for _, resp := range responses {
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
	return results, nil
}


var _ queryconnect.QueryServiceClient = &InlineQueryClient{}

// InlineQueryClient is used to perform queries by the unit test harness.
// It executes the query server logic in-process and does not require network communication.
type InlineQueryClient struct {
	service   *Service
	collector *responseCollector
}

func NewInlineQueryClient(service *Service) *InlineQueryClient {
	return &InlineQueryClient{
		service:   service,
		collector: &responseCollector{responses: []*querypb.ExecuteQueryResponse{}},
	}
}

func (c *InlineQueryClient) drainResponses() []*querypb.ExecuteQueryResponse {
	responses := c.collector.responses
	c.collector.responses = []*querypb.ExecuteQueryResponse{}
	return responses
}

func (c *InlineQueryClient) ExecuteQuery(
	ctx context.Context,
	req *connect.Request[querypb.ExecuteQueryRequest],
) (*connect.ServerStreamForClient[querypb.ExecuteQueryResponse], error) {
	if c.service == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query service not initialized"))
	}
	if err := c.service.ExecuteQueryInternal(ctx, req, c.collector); err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	return nil, nil
}

func (c *InlineQueryClient) BeginTransaction(
	ctx context.Context,
	req *connect.Request[querypb.BeginTransactionRequest],
) (*connect.Response[querypb.BeginTransactionResponse], error) {
	if c.service == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query service not initialized"))
	}
	return c.service.BeginTransaction(ctx, req)
}

func (c *InlineQueryClient) CommitTransaction(
	ctx context.Context,
	req *connect.Request[querypb.CommitTransactionRequest],
) (*connect.Response[querypb.CommitTransactionResponse], error) {
	if c.service == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query service not initialized"))
	}
	return c.service.CommitTransaction(ctx, req)
}

func (c *InlineQueryClient) RollbackTransaction(
	ctx context.Context,
	req *connect.Request[querypb.RollbackTransactionRequest],
) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	if c.service == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query service not initialized"))
	}
	return c.service.RollbackTransaction(ctx, req)
}

func (c *InlineQueryClient) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query connection should not be pinged directly"))
}

type responseCollector struct {
	responses []*querypb.ExecuteQueryResponse
}

func (r *responseCollector) Send(resp *querypb.ExecuteQueryResponse) error {
	r.responses = append(r.responses, resp)
	return nil
}

func (r *responseCollector) Receive() bool {
	if len(r.responses) == 0 {
		return false
	}
	r.responses = r.responses[1:]
	return true
}

func (r *responseCollector) Msg() *querypb.ExecuteQueryResponse {
	if len(r.responses) == 0 {
		return nil
	}
	return r.responses[0]
}
