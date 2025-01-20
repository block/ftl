package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/puzpuzpuz/xsync/v3"
	"google.golang.org/protobuf/types/known/timestamppb"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
)

// MultiService handles multiple database instances
type MultiService struct {
	services *xsync.MapOf[schema.RefKey, *Service]
}

func NewMultiService() *MultiService {
	return &MultiService{services: xsync.NewMapOf[schema.RefKey, *Service]()}
}

func (m *MultiService) AddService(module string, name string, svc *Service) error {
	refKey := schema.RefKey{Module: module, Name: name}
	m.services.Store(refKey, svc)
	envVar := strings.ToUpper("FTL_QUERY_ADDRESS_" + name)
	if err := os.Setenv(envVar, svc.endpoint.String()); err != nil {
		return fmt.Errorf("failed to set query address for %s: %w", name, err)
	}
	return nil
}

func (m *MultiService) GetServices() map[schema.RefKey]*Service {
	result := make(map[schema.RefKey]*Service)
	m.services.Range(func(key schema.RefKey, value *Service) bool {
		result[key] = value
		return true
	})
	return result
}

func (m *MultiService) Close() error {
	var lastErr error
	m.services.Range(func(key schema.RefKey, svc *Service) bool {
		if err := svc.Close(); err != nil {
			lastErr = err
		}
		return true
	})
	return lastErr
}

func (m *MultiService) GetService(ref schema.RefKey) (*Service, bool) {
	return m.services.Load(ref)
}

type Config struct {
	DSN      string
	Engine   string
	Endpoint *url.URL
}

func (c *Config) Validate() error {
	if c.DSN == "" {
		return fmt.Errorf("database connection string is required")
	}
	if c.Endpoint == nil {
		return fmt.Errorf("service endpoint is required")
	}
	switch c.Engine {
	case "mysql", "postgres":
		return nil
	default:
		return fmt.Errorf("unsupported database engine: %s (supported: mysql, postgres)", c.Engine)
	}
}

type Service struct {
	endpoint     *url.URL
	lock         sync.RWMutex
	transactions map[string]*sql.Tx
	db           *sql.DB
	engine       string
}

// DB represents a database that can execute queries
type DB interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func New(ctx context.Context, config Config) (*Service, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).Scope("query")
	ctx = log.ContextWithLogger(ctx, logger)

	db, err := sql.Open(getDriverName(config.Engine), config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Service{
		endpoint:     config.Endpoint,
		transactions: make(map[string]*sql.Tx),
		db:           db,
		engine:       config.Engine,
	}, nil
}

func (s *Service) Close() error {
	if s.db != nil {
		s.db.Close()
	}
	return nil
}

var _ queryconnect.QueryServiceHandler = (*Service)(nil)

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) BeginTransaction(ctx context.Context, req *connect.Request[querypb.BeginTransactionRequest]) (*connect.Response[querypb.BeginTransactionResponse], error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to begin transaction: %w", err))
	}

	transactionID := uuid.New().String()

	s.lock.Lock()
	s.transactions[transactionID] = tx
	s.lock.Unlock()

	return connect.NewResponse(&querypb.BeginTransactionResponse{
		TransactionId: transactionID,
		Status:        querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *Service) CommitTransaction(ctx context.Context, req *connect.Request[querypb.CommitTransactionRequest]) (*connect.Response[querypb.CommitTransactionResponse], error) {
	s.lock.Lock()
	tx, exists := s.transactions[req.Msg.TransactionId]
	if !exists {
		s.lock.Unlock()
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("transaction %s not found", req.Msg.TransactionId))
	}
	delete(s.transactions, req.Msg.TransactionId)
	s.lock.Unlock()

	if err := tx.Commit(); err != nil {
		return connect.NewResponse(&querypb.CommitTransactionResponse{
			Status: querypb.TransactionStatus_TRANSACTION_STATUS_FAILED,
		}), nil
	}

	return connect.NewResponse(&querypb.CommitTransactionResponse{
		Status: querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *Service) RollbackTransaction(ctx context.Context, req *connect.Request[querypb.RollbackTransactionRequest]) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	s.lock.Lock()
	tx, exists := s.transactions[req.Msg.TransactionId]
	if !exists {
		s.lock.Unlock()
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("transaction %s not found", req.Msg.TransactionId))
	}
	delete(s.transactions, req.Msg.TransactionId)
	s.lock.Unlock()

	if err := tx.Rollback(); err != nil {
		return connect.NewResponse(&querypb.RollbackTransactionResponse{
			Status: querypb.TransactionStatus_TRANSACTION_STATUS_FAILED,
		}), nil
	}

	return connect.NewResponse(&querypb.RollbackTransactionResponse{
		Status: querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *Service) ExecuteQuery(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream *connect.ServerStream[querypb.ExecuteQueryResponse]) error {
	if req.Msg.TransactionId != nil && *req.Msg.TransactionId != "" {
		s.lock.RLock()
		tx, ok := s.transactions[*req.Msg.TransactionId]
		s.lock.RUnlock()
		if !ok {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("transaction not found: %s", *req.Msg.TransactionId))
		}
		return s.executeQuery(ctx, tx, req.Msg, stream)
	}
	return s.executeQuery(ctx, s.db, req.Msg, stream)
}

// SetEndpoint sets the endpoint after the server is confirmed ready
func (s *Service) SetEndpoint(endpoint *url.URL) {
	s.endpoint = endpoint
}

func (s *Service) executeQuery(ctx context.Context, db DB, req *querypb.ExecuteQueryRequest, stream *connect.ServerStream[querypb.ExecuteQueryResponse]) error {
	switch req.CommandType {
	case querypb.CommandType_COMMAND_TYPE_EXEC:
		result, err := db.ExecContext(ctx, req.RawSql, convertParameters(req.Parameters)...)
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to execute query: %w", err))
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get rows affected: %w", err))
		}
		lastInsertID, err := result.LastInsertId()
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get last insert id: %w", err))
		}
		err = stream.Send(&querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_ExecResult{
				ExecResult: &querypb.ExecResult{
					RowsAffected: rowsAffected,
					LastInsertId: &lastInsertID,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to send exec result: %w", err)
		}

	case querypb.CommandType_COMMAND_TYPE_ONE:
		row := db.QueryRowContext(ctx, req.RawSql, convertParameters(req.Parameters)...)
		result, err := scanRowToMap(row, req.ResultColumns)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return handleNoRows(stream)
			}
			return fmt.Errorf("failed to scan row: %w", err)
		}

		err = stream.Send(&querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_RowResults{
				RowResults: &querypb.RowResults{
					Rows: result,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to send row results: %w", err)
		}

	case querypb.CommandType_COMMAND_TYPE_MANY:
		rows, err := db.QueryContext(ctx, req.RawSql, convertParameters(req.Parameters)...)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return handleNoRows(stream)
			}
			return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to execute query: %w", err))
		}
		defer rows.Close()

		rowCount := int32(0)
		batchSize := int32(100) // Default batch size
		if req.BatchSize != nil {
			batchSize = *req.BatchSize
		}

		for rows.Next() {
			result, err := scanRowToMap(rows, req.ResultColumns)
			if err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to scan row: %w", err))
			}

			rowCount++
			hasMore := rowCount < batchSize
			if err := stream.Send(&querypb.ExecuteQueryResponse{
				Result: &querypb.ExecuteQueryResponse_RowResults{
					RowResults: &querypb.RowResults{
						Rows:    result,
						HasMore: hasMore,
					},
				},
			}); err != nil {
				return fmt.Errorf("failed to send row results: %w", err)
			}

			if !hasMore {
				rowCount = 0
			}
		}
	case querypb.CommandType_COMMAND_TYPE_UNSPECIFIED:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown command type: %v", req.CommandType))
	}
	return nil
}

// scanRowToMap scans a row into a map[string]*querypb.SQLValue where keys are column names and values are the row values
func scanRowToMap(row any, resultColumns []string) (map[string]*querypb.SQLValue, error) {
	if len(resultColumns) == 0 {
		return nil, fmt.Errorf("result_columns required for scanning rows")
	}

	// Get column names from the row
	var dbColumns []string
	switch r := row.(type) {
	case *sql.Rows:
		var err error
		dbColumns, err = r.Columns()
		if err != nil {
			return nil, fmt.Errorf("failed to get column names: %w", err)
		}
	case *sql.Row:
		// For sql.Row we can't get column names, but we know they must match our query
		dbColumns = resultColumns
	default:
		return nil, fmt.Errorf("unsupported row type: %T", row)
	}

	if len(dbColumns) != len(resultColumns) {
		return nil, fmt.Errorf("column count mismatch: got %d columns from DB but expected %d columns", len(dbColumns), len(resultColumns))
	}

	// Create value slices for scanning
	values := make([]any, len(dbColumns))
	valuePointers := make([]any, len(dbColumns))
	for i := range values {
		valuePointers[i] = &values[i]
	}

	var err error
	switch r := row.(type) {
	case *sql.Row:
		err = r.Scan(valuePointers...)
	case *sql.Rows:
		err = r.Scan(valuePointers...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := make(map[string]*querypb.SQLValue)
	for i, val := range values {
		col := dbColumns[i]
		if val == nil {
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_NullValue{NullValue: true}}
			continue
		}

		switch v := val.(type) {
		case []byte:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_BytesValue{BytesValue: v}}
		case string:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_StringValue{StringValue: v}}
		case int64:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: v}}
		case uint64:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: int64(v)}}
		case float64:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_FloatValue{FloatValue: v}}
		case bool:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_BoolValue{BoolValue: v}}
		case time.Time:
			result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_TimestampValue{TimestampValue: timestamppb.New(v)}}
		default:
			// Try to convert numeric types
			if iv, ok := val.(int); ok {
				result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: int64(iv)}}
			} else if iv32, ok := val.(int32); ok {
				result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: int64(iv32)}}
			} else if uv, ok := val.(uint); ok {
				result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: int64(uv)}}
			} else if uv32, ok := val.(uint32); ok {
				result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_IntValue{IntValue: int64(uv32)}}
			} else if fv32, ok := val.(float32); ok {
				result[col] = &querypb.SQLValue{Value: &querypb.SQLValue_FloatValue{FloatValue: float64(fv32)}}
			} else {
				return nil, fmt.Errorf("unsupported value type for column %s: %T with value %v", col, val, val)
			}
		}
	}

	return result, nil
}

func convertParameters(params []*querypb.SQLValue) []interface{} {
	if params == nil {
		return nil
	}
	result := make([]interface{}, len(params))
	for i, p := range params {
		if p == nil {
			result[i] = nil
			continue
		}
		switch v := p.Value.(type) {
		case *querypb.SQLValue_StringValue:
			result[i] = v.StringValue
		case *querypb.SQLValue_IntValue:
			result[i] = v.IntValue
		case *querypb.SQLValue_FloatValue:
			result[i] = v.FloatValue
		case *querypb.SQLValue_BoolValue:
			result[i] = v.BoolValue
		case *querypb.SQLValue_BytesValue:
			result[i] = v.BytesValue
		case *querypb.SQLValue_TimestampValue:
			result[i] = v.TimestampValue.AsTime()
		case *querypb.SQLValue_NullValue:
			result[i] = nil
		default:
			result[i] = nil
		}
	}
	return result
}

func getDriverName(engine string) string {
	switch engine {
	case "postgres":
		return "postgres"
	default:
		return "mysql"
	}
}

func handleNoRows(stream *connect.ServerStream[querypb.ExecuteQueryResponse]) error {
	err := stream.Send(&querypb.ExecuteQueryResponse{
		Result: &querypb.ExecuteQueryResponse_RowResults{
			RowResults: &querypb.RowResults{
				Rows:    make(map[string]*querypb.SQLValue),
				HasMore: false,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send no rows response: %w", err)
	}
	return nil
}
