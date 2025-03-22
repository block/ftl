package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/puzpuzpuz/xsync/v3"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/log"
)

var _ queryconnect.QueryServiceHandler = (*Service)(nil)

// Service proxies query requests to multiple database instances
type Service struct {
	// Maps database name to connection
	conns *xsync.MapOf[string, *queryConn]
	// Mutex for service operations
	mu sync.Mutex
}

func New(ctx context.Context, module *schema.Module, addresses *xsync.MapOf[string, string]) (*Service, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Initializing query service for module %s", module.Name)

	s := &Service{
		conns: xsync.NewMapOf[string, *queryConn](),
	}

	// Initialize connections for all databases in the module
	if err := s.UpdateConnections(ctx, module, addresses); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) AddQueryConn(ctx context.Context, name string, dsn deploymentcontext.Database) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var engine string
	switch dsn.DBType {
	case deploymentcontext.DBTypePostgres:
		engine = "postgres"
	case deploymentcontext.DBTypeMySQL:
		engine = "mysql"
	default:
		return fmt.Errorf("unsupported database type: %s", dsn.DBType)
	}
	svc, err := newQueryConn(ctx, dsn.DSN, engine)
	if err != nil {
		return err
	}
	s.conns.Store(name, svc)
	return nil
}

func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error
	s.conns.Range(func(key string, svc *queryConn) bool {
		if err := svc.Close(); err != nil {
			lastErr = err
		}
		return true
	})
	return lastErr
}

func (s *Service) BeginTransaction(ctx context.Context, req *connect.Request[querypb.BeginTransactionRequest]) (*connect.Response[querypb.BeginTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, err
	}
	return conn.BeginTransaction(ctx, req)
}

func (s *Service) CommitTransaction(ctx context.Context, req *connect.Request[querypb.CommitTransactionRequest]) (*connect.Response[querypb.CommitTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, err
	}
	return conn.CommitTransaction(ctx, req)
}

// serverStream is a stream of query responses implemented by *connect.ServerStream[querypb.ExecuteQueryResponse]
// It's an interface for testing purposes.
type serverStream interface {
	Send(resp *querypb.ExecuteQueryResponse) error
}

func (s *Service) ExecuteQuery(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream *connect.ServerStream[querypb.ExecuteQueryResponse]) error {
	return s.ExecuteQueryInternal(ctx, req, stream)
}

func (s *Service) ExecuteQueryInternal(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream serverStream) error {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return err
	}
	return conn.ExecuteQuery(ctx, req, stream)
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) RollbackTransaction(ctx context.Context, req *connect.Request[querypb.RollbackTransactionRequest]) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, err
	}
	return conn.RollbackTransaction(ctx, req)
}

func (s *Service) getConn(name string) (*queryConn, bool) {
	return s.conns.Load(name)
}

func (s *Service) getConnOrError(name string) (*queryConn, error) {
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("database name is required"))
	}
	conn, ok := s.getConn(name)
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("database connection for %s not found", name))
	}
	return conn, nil
}

// UpdateConnections updates the connections based on a new module and addresses during hot reloading.
// It fails if any database has configuration or connection issues.
func (s *Service) UpdateConnections(ctx context.Context, module *schema.Module, addresses *xsync.MapOf[string, string]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	logger := log.FromContext(ctx)
	logger.Debugf("Updating database connections for module %s", module.Name)

	newDatabases := make(map[string]bool)

	// Initialize new connections and update existing ones
	for _, decl := range module.Decls {
		db, ok := decl.(*schema.Database)
		if !ok {
			continue
		}

		newDatabases[db.Name] = true

		dbadr, ok := addresses.Load(db.Name)
		if !ok {
			return fmt.Errorf("could not find DSN for database %s", db.Name)
		}

		// If we already have a connection for this database, reuse it
		if _, exists := s.conns.Load(db.Name); exists {
			logger.Debugf("Reusing existing connection for database %s", db.Name)
			continue
		}

		logger.Debugf("Creating new connection for database %s", db.Name)
		parts := strings.Split(dbadr, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf("failed to parse port for database %s: %w", db.Name, err)
		}
		host := parts[0]
		var sdsn string

		switch db.Type {
		case schema.MySQLDatabaseType:
			sdsn = dsn.MySQLDSN(db.Name, dsn.Host(host), dsn.Port(port))
		case schema.PostgresDatabaseType:
			sdsn = dsn.PostgresDSN(db.Name, dsn.Host(host), dsn.Port(port))
		default:
			logger.Debugf("Unsupported database type: %T", db.Type)
			return fmt.Errorf("unsupported database type for %s: %s", db.Name, db.Type)
		}

		querySvc, err := newQueryConn(ctx, sdsn, db.Type)
		if err != nil {
			logger.Debugf("Failed to create query service for database %s: %v", db.Name, err)
			return fmt.Errorf("failed to create query service for database %s: %w", db.Name, err)
		}

		s.conns.Store(db.Name, querySvc)
		logger.Debugf("Successfully stored query connection for database %s", db.Name)
	}

	// Close connections for databases that no longer exist in the module
	var databasesToRemove []string
	s.conns.Range(func(dbName string, conn *queryConn) bool {
		if !newDatabases[dbName] {
			databasesToRemove = append(databasesToRemove, dbName)
		}
		return true
	})

	for _, dbName := range databasesToRemove {
		if conn, exists := s.conns.Load(dbName); exists {
			logger.Debugf("Closing connection for removed database %s", dbName)
			if err := conn.Close(); err != nil {
				logger.Debugf("Error closing connection for database %s: %v", dbName, err)
				// don't error so we can continue closing other connections
			}
			s.conns.Delete(dbName)
		}
	}

	return nil
}

type queryConn struct {
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

func newQueryConn(ctx context.Context, dsn string, engine string) (*queryConn, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database connection string is required")
	}

	switch engine {
	case "mysql", "postgres":
	default:
		return nil, fmt.Errorf("unsupported database engine: %s (supported: mysql, postgres)", engine)
	}

	logger := log.FromContext(ctx).Scope("query")
	ctx = log.ContextWithLogger(ctx, logger)

	db, err := sql.Open(getDriverName(engine), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &queryConn{
		transactions: make(map[string]*sql.Tx),
		db:           db,
		engine:       engine,
	}, nil
}

func (s *queryConn) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Rollback any open transactions
	for id, tx := range s.transactions {
		_ = tx.Rollback() //nolint:errcheck
		delete(s.transactions, id)
	}

	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}
	return nil
}

func (s *queryConn) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("query connection should not be pinged directly"))
}

func (s *queryConn) BeginTransaction(ctx context.Context, req *connect.Request[querypb.BeginTransactionRequest]) (*connect.Response[querypb.BeginTransactionResponse], error) {
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

func (s *queryConn) CommitTransaction(ctx context.Context, req *connect.Request[querypb.CommitTransactionRequest]) (*connect.Response[querypb.CommitTransactionResponse], error) {
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

func (s *queryConn) RollbackTransaction(ctx context.Context, req *connect.Request[querypb.RollbackTransactionRequest]) (*connect.Response[querypb.RollbackTransactionResponse], error) {
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

func (s *queryConn) ExecuteQuery(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream serverStream) error {
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

func (s *queryConn) executeQuery(ctx context.Context, db DB, req *querypb.ExecuteQueryRequest, stream serverStream) error {
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

	exportName := func(name string) string {
		return strings.ToUpper(name[:1]) + name[1:]
	}

	// create a result struct which will be encoded to JSON
	structFields := make([]reflect.StructField, len(resultColumns))
	for i, col := range resultColumns {
		structFields[i] = reflect.StructField{
			Name: exportName(col.TypeName),
			Type: reflect.TypeFor[any](),
			Tag:  reflect.StructTag(`json:"` + col.TypeName + `"`),
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

		field := structValue.FieldByName(exportName(typeName))
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
		return "pgx"
	case "mysql":
		return "mysql"
	default:
		panic("unsupported database engine: " + engine)
	}
}

func handleNoRows(stream serverStream) error {
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
