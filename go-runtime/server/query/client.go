package query

import (
	"context"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/tuple"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	iquery "github.com/block/ftl/backend/runner/query"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/go-runtime/internal"
	"github.com/block/ftl/go-runtime/server/rpccontext"
)

const (
	transactionMetadataString = "ftl-transaction"
	TransactionMetadataKey    = internal.MetadataKey(transactionMetadataString)
)

func GetTransactionMetadata(txnID string) *ftlv1.Metadata_Pair {
	return &ftlv1.Metadata_Pair{
		Key:   transactionMetadataString,
		Value: txnID,
	}
}

func BeginTransaction(ctx context.Context, dbName string) (string, error) {
	client := rpccontext.ClientFromContext[queryconnect.QueryServiceClient](ctx)
	resp, err := client.BeginTransaction(ctx, connect.NewRequest(&querypb.BeginTransactionRequest{
		DatabaseName: dbName,
	}))
	if err != nil {
		return "", errors.Wrap(err, "failed to begin transaction")
	}
	if resp.Msg.Status != querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS {
		return "", errors.Errorf("transaction on %s was not started", dbName)
	}
	return resp.Msg.TransactionId, nil
}

func RollbackCurrentTransaction(ctx context.Context, dbName string) error {
	client := rpccontext.ClientFromContext[queryconnect.QueryServiceClient](ctx)
	txID, ok := internal.CallMetadataFromContext(ctx)[TransactionMetadataKey]
	if !ok {
		return errors.Errorf("no active transaction found")
	}
	resp, err := client.RollbackTransaction(ctx, connect.NewRequest(&querypb.RollbackTransactionRequest{
		DatabaseName:  dbName,
		TransactionId: txID,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to rollback transaction")
	}
	if resp.Msg.Status != querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS {
		return errors.Errorf("transaction %s was not rolled back", txID)
	}
	return nil
}

func CommitCurrentTransaction(ctx context.Context, dbName string) error {
	client := rpccontext.ClientFromContext[queryconnect.QueryServiceClient](ctx)
	txID, ok := internal.CallMetadataFromContext(ctx)[TransactionMetadataKey]
	if !ok {
		return errors.Errorf("no active transaction found")
	}
	resp, err := client.CommitTransaction(ctx, connect.NewRequest(&querypb.CommitTransactionRequest{
		DatabaseName:  dbName,
		TransactionId: txID,
	}))
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}
	if resp.Msg.Status != querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS {
		return errors.Errorf("transaction %s was not committed", txID)
	}
	return nil
}

func One[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) (resp Resp, err error) {
	results, err := performQuery[Resp](ctx, dbName, querypb.CommandType_COMMAND_TYPE_ONE, rawSQL, params, colFieldNames)
	if err != nil {
		return resp, errors.WithStack(err)
	}
	if len(results) == 0 {
		return resp, errors.Errorf("no results found")
	}
	return results[0], nil
}

func Many[Req, Resp any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]Resp, error) {
	return errors.WithStack2(performQuery[Resp](ctx, dbName, querypb.CommandType_COMMAND_TYPE_MANY, rawSQL, params, colFieldNames))
}

func Exec[Req any](ctx context.Context, dbName string, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) error {
	_, err := performQuery[any](ctx, dbName, querypb.CommandType_COMMAND_TYPE_EXEC, rawSQL, params, colFieldNames)
	return errors.WithStack(err)
}

// performQuery performs a SQL query and returns the results.
//
// note: accepts colFieldNames as []tuple.Pair[string, string] rather than map[string]string to preserve column order
func performQuery[T any](ctx context.Context, dbName string, commandType querypb.CommandType, rawSQL string, params []any, colFieldNames []tuple.Pair[string, string]) ([]T, error) {
	client := rpccontext.ClientFromContext[queryconnect.QueryServiceClient](ctx)

	paramsJSON, err := encoding.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal parameters")
	}

	resultCols := make([]*querypb.ResultColumn, 0, len(colFieldNames))
	for _, pair := range colFieldNames {
		resultCols = append(resultCols, &querypb.ResultColumn{
			SqlName:  pair.A,
			TypeName: pair.B,
		})
	}

	req := &querypb.ExecuteQueryRequest{
		DatabaseName:   dbName,
		RawSql:         rawSQL,
		CommandType:    commandType,
		ParametersJson: string(paramsJSON),
		ResultColumns:  resultCols,
	}
	if md, ok := internal.MaybeCallMetadataFromContext(ctx).Get(); ok {
		if txID, ok := md[TransactionMetadataKey]; ok {
			req.TransactionId = &txID
		}
	}
	stream, err := client.ExecuteQuery(ctx, connect.NewRequest(req))
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
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
			return nil, errors.Wrap(err, "failed to receive query results")
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
				return nil, errors.Wrap(err, "failed to unmarshal row data")
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
	service   *iquery.Service
	collector *responseCollector
}

func NewInlineQueryClient(service *iquery.Service) *InlineQueryClient {
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
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query service not initialized")))
	}
	if err := c.service.ExecuteQueryInternal(ctx, req, c.collector); err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	return nil, nil
}

func (c *InlineQueryClient) BeginTransaction(
	ctx context.Context,
	req *connect.Request[querypb.BeginTransactionRequest],
) (*connect.Response[querypb.BeginTransactionResponse], error) {
	if c.service == nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query service not initialized")))
	}
	return errors.WithStack2(c.service.BeginTransaction(ctx, req))
}

func (c *InlineQueryClient) CommitTransaction(
	ctx context.Context,
	req *connect.Request[querypb.CommitTransactionRequest],
) (*connect.Response[querypb.CommitTransactionResponse], error) {
	if c.service == nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query service not initialized")))
	}
	return errors.WithStack2(c.service.CommitTransaction(ctx, req))
}

func (c *InlineQueryClient) RollbackTransaction(
	ctx context.Context,
	req *connect.Request[querypb.RollbackTransactionRequest],
) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	if c.service == nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query service not initialized")))
	}
	return errors.WithStack2(c.service.RollbackTransaction(ctx, req))
}

func (c *InlineQueryClient) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query connection should not be pinged directly")))
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
