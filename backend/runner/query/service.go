package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/puzpuzpuz/xsync/v3"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
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
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
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
	params, err := parseJSONParameters(req.GetParametersJson())
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to parse parameters: %w", err))
	}

	switch req.CommandType {
	case querypb.CommandType_COMMAND_TYPE_EXEC:
		result, err := db.ExecContext(ctx, req.RawSql, params...)
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

		protoResp := &querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_ExecResult{
				ExecResult: &querypb.ExecResult{
					RowsAffected: rowsAffected,
					LastInsertId: &lastInsertID,
				},
			},
		}

		err = stream.Send(protoResp)
		if err != nil {
			return fmt.Errorf("failed to send exec result: %w", err)
		}

	case querypb.CommandType_COMMAND_TYPE_ONE:
		row := db.QueryRowContext(ctx, req.RawSql, params...)
		jsonRows, err := scanRowToMap(row, req.ResultColumns)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return handleNoRows(stream)
			}
			return fmt.Errorf("failed to scan row: %w", err)
		}

		protoResp := &querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_RowResults{
				RowResults: &querypb.RowResults{
					JsonRows: jsonRows,
					HasMore:  false,
				},
			},
		}

		err = stream.Send(protoResp)
		if err != nil {
			return fmt.Errorf("failed to send row results: %w", err)
		}

	case querypb.CommandType_COMMAND_TYPE_MANY:
		rows, err := db.QueryContext(ctx, req.RawSql, params...)
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
			jsonRows, err := scanRowToMap(rows, req.ResultColumns)
			if err != nil {
				return connect.NewError(connect.CodeInternal, fmt.Errorf("failed to scan row: %w", err))
			}

			rowCount++
			hasMore := rowCount < batchSize

			protoResp := &querypb.ExecuteQueryResponse{
				Result: &querypb.ExecuteQueryResponse_RowResults{
					RowResults: &querypb.RowResults{
						JsonRows: jsonRows,
						HasMore:  hasMore,
					},
				},
			}

			if err := stream.Send(protoResp); err != nil {
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

func parseJSONParameters(paramsJSON string) ([]any, error) {
	if paramsJSON == "" {
		return nil, nil
	}

	var params []any
	if err := encoding.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
	}

	for i, param := range params {
		if str, ok := param.(string); ok {
			// Convert string to time.Time if it matches a known format
			if t, err := parseTimeString(str); err == nil {
				params[i] = t
			}
		}
	}

	return params, nil
}

// parseTimeString attempts to parse a time string in various formats
func parseTimeString(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,     // 2024-01-01T12:00:00Z
		time.RFC3339Nano, // 2024-01-01T12:00:00.000Z
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse time string: %s", s)
}

// scanRowToMap scans a row and returns a JSON string representation
func scanRowToMap(row any, resultColumns []*querypb.ResultColumn) (string, error) {
	if len(resultColumns) == 0 {
		return "", fmt.Errorf("result_columns required for scanning rows")
	}

	typeNameBySQLName := make(map[string]string)
	sqlColumns := make([]string, 0, len(resultColumns))
	for _, col := range resultColumns {
		sqlColumns = append(sqlColumns, col.SqlName)
		typeNameBySQLName[col.SqlName] = col.TypeName
	}

	// Get column names from the row
	var dbColumns []string
	switch r := row.(type) {
	case *sql.Rows:
		var err error
		dbColumns, err = r.Columns()
		if err != nil {
			return "", fmt.Errorf("failed to get column names: %w", err)
		}
	case *sql.Row:
		// For sql.Row we can't get column names, but we know they must match our query
		dbColumns = sqlColumns
	default:
		return "", fmt.Errorf("unsupported row type: %T", row)
	}

	if len(dbColumns) != len(resultColumns) {
		return "", fmt.Errorf("column count mismatch: got %d columns from DB but expected %d columns", len(dbColumns), len(resultColumns))
	}

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
		return "", fmt.Errorf("failed to scan row: %w", err)
	}

	// create a result struct which will be encoded to JSON
	structFields := make([]reflect.StructField, len(resultColumns))
	for i, col := range resultColumns {
		structFields[i] = reflect.StructField{
			Name: col.TypeName,
			Type: reflect.TypeFor[any](),
		}
	}

	structType := reflect.StructOf(structFields)
	structValue := reflect.New(structType).Elem()
	for i, val := range values {
		typeName := typeNameBySQLName[dbColumns[i]]
		if val == nil {
			continue
		}

		var fieldValue any
		switch v := val.(type) {
		case []byte:
			str := string(v)
			if t, err := parseTimeString(str); err == nil {
				fieldValue = t
			} else {
				fieldValue = str
			}
		case string:
			if t, err := parseTimeString(v); err == nil {
				fieldValue = t
			} else {
				fieldValue = v
			}
		default:
			fieldValue = v
		}

		field := structValue.FieldByName(typeName)
		if field.IsValid() {
			field.Set(reflect.ValueOf(fieldValue))
		}
	}

	jsonBytes, err := encoding.Marshal(structValue.Interface())
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonBytes), nil
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
				JsonRows: "",
				HasMore:  false,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send no rows response: %w", err)
	}
	return nil
}
