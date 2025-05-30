package query

import (
	"context"
	"database/sql"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	errors "github.com/alecthomas/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/puzpuzpuz/xsync/v3"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	queryconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1/querypbconnect"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/common/encoding"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/dsn"
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
		return nil, errors.WithStack(err)
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
		return errors.Errorf("unsupported database type: %s", dsn.DBType)
	}
	svc, err := newQueryConn(ctx, dsn.DSN, engine)
	if err != nil {
		return errors.WithStack(err)
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
	return errors.WithStack(lastErr)
}

func (s *Service) BeginTransaction(ctx context.Context, req *connect.Request[querypb.BeginTransactionRequest]) (*connect.Response[querypb.BeginTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(conn.BeginTransaction(ctx, req))
}

func (s *Service) CommitTransaction(ctx context.Context, req *connect.Request[querypb.CommitTransactionRequest]) (*connect.Response[querypb.CommitTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(conn.CommitTransaction(ctx, req))
}

// serverStream is a stream of query responses implemented by *connect.ServerStream[querypb.ExecuteQueryResponse]
// It's an interface for testing purposes.
type serverStream interface {
	Send(resp *querypb.ExecuteQueryResponse) error
}

func (s *Service) ExecuteQuery(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream *connect.ServerStream[querypb.ExecuteQueryResponse]) error {
	return errors.WithStack(s.ExecuteQueryInternal(ctx, req, stream))
}

func (s *Service) ExecuteQueryInternal(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream serverStream) error {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(conn.ExecuteQuery(ctx, req, stream))
}

func (s *Service) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

func (s *Service) RollbackTransaction(ctx context.Context, req *connect.Request[querypb.RollbackTransactionRequest]) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	conn, err := s.getConnOrError(req.Msg.DatabaseName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return errors.WithStack2(conn.RollbackTransaction(ctx, req))
}

func (s *Service) getConn(name string) (*queryConn, bool) {
	return s.conns.Load(name)
}

func (s *Service) getConnOrError(name string) (*queryConn, error) {
	if name == "" {
		return nil, errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Errorf("database name is required")))
	}
	conn, ok := s.getConn(name)
	if !ok {
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("database connection for %s not found", name)))
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
			return errors.Errorf("could not find DSN for database %s", db.Name)
		}

		// If we already have a connection for this database, reuse it
		if _, exists := s.conns.Load(db.Name); exists {
			logger.Debugf("Reusing existing connection for database %s", db.Name)
			continue
		}

		parts := strings.Split(dbadr, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return errors.Wrapf(err, "failed to parse port for database %s", db.Name)
		}
		host := parts[0]
		var sdsn string

		logger.Debugf("Creating new connection for database %s to %s:%d", db.Name, host, port)
		switch db.Type {
		case schema.MySQLDatabaseType:
			sdsn = dsn.MySQLDSN(db.Name, dsn.Host(host), dsn.Port(port))
		case schema.PostgresDatabaseType:
			sdsn = dsn.PostgresDSN(db.Name, dsn.Host(host), dsn.Port(port))
		default:
			logger.Debugf("Unsupported database type: %T", db.Type)
			return errors.Errorf("unsupported database type for %s: %s", db.Name, db.Type)
		}

		querySvc, err := newQueryConn(ctx, sdsn, db.Type)
		if err != nil {
			logger.Debugf("Failed to create query service for database %s: %v", db.Name, err)
			return errors.Wrapf(err, "failed to create query service for database %s", db.Name)
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
	transactions map[string]*txWrapper
	db           *sql.DB
	engine       string
}

// txWrapper holds both the transaction and its dedicated connection
type txWrapper struct {
	tx         *sql.Tx
	conn       *sql.Conn
	cancelFunc context.CancelFunc //nolint:forbidigo
}

func (t *txWrapper) cancel() {
	t.cancelFunc()
}

// DB represents a database that can execute queries
type DB interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func newQueryConn(ctx context.Context, dsn string, engine string) (*queryConn, error) {
	db, err := sql.Open(getDriverName(engine), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return &queryConn{
		transactions: make(map[string]*txWrapper),
		db:           db,
		engine:       engine,
	}, nil
}

func (s *queryConn) Close() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Rollback any open transactions and close their connections
	for id, wrapper := range s.transactions {
		wrapper.cancel()
		delete(s.transactions, id)
	}

	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close database")
		}
	}
	return nil
}

func (s *queryConn) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Errorf("query connection should not be pinged directly")))
}

func (s *queryConn) BeginTransaction(ctx context.Context, req *connect.Request[querypb.BeginTransactionRequest]) (*connect.Response[querypb.BeginTransactionResponse], error) {
	// use context.Background() instead of the current request context. The transaction lifecycle
	// must be managed independently to avoid premature cancelation; it may extend beyond the life of
	// the current request if multiple verbs are executed in a single transaction.
	txCtx, cancel := context.WithTimeoutCause(context.Background(), 30*time.Second, errors.Errorf("transaction timed out")) // TODO: configure txn timeouts via db/config.toml
	conn, err := s.db.Conn(txCtx)
	if err != nil {
		cancel()
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to get dedicated connection")))
	}

	tx, err := conn.BeginTx(txCtx, nil)
	if err != nil {
		cancel()
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to begin transaction")))
	}

	txID := uuid.NewString()
	s.lock.Lock()
	if s.transactions == nil {
		s.transactions = make(map[string]*txWrapper)
	}
	s.transactions[txID] = &txWrapper{
		tx:         tx,
		conn:       conn,
		cancelFunc: cancel,
	}
	s.lock.Unlock()

	return connect.NewResponse(&querypb.BeginTransactionResponse{
		TransactionId: txID,
		Status:        querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *queryConn) CommitTransaction(ctx context.Context, req *connect.Request[querypb.CommitTransactionRequest]) (*connect.Response[querypb.CommitTransactionResponse], error) {
	s.lock.Lock()
	wrapper, exists := s.transactions[req.Msg.GetTransactionId()]
	if !exists {
		s.lock.Unlock()
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("transaction %s not found", req.Msg.TransactionId)))
	}
	delete(s.transactions, req.Msg.TransactionId)
	s.lock.Unlock()

	defer wrapper.cancel()
	if err := wrapper.tx.Commit(); err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to commit transaction")))
	}
	return connect.NewResponse(&querypb.CommitTransactionResponse{
		Status: querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *queryConn) RollbackTransaction(ctx context.Context, req *connect.Request[querypb.RollbackTransactionRequest]) (*connect.Response[querypb.RollbackTransactionResponse], error) {
	s.lock.Lock()
	wrapper, exists := s.transactions[req.Msg.GetTransactionId()]
	if !exists {
		s.lock.Unlock()
		return nil, errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("transaction %s not found", req.Msg.TransactionId)))
	}
	delete(s.transactions, req.Msg.TransactionId)
	s.lock.Unlock()

	defer wrapper.cancel()
	if err := wrapper.tx.Rollback(); err != nil {
		return nil, errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to rollback transaction")))
	}
	return connect.NewResponse(&querypb.RollbackTransactionResponse{
		Status: querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS,
	}), nil
}

func (s *queryConn) ExecuteQuery(ctx context.Context, req *connect.Request[querypb.ExecuteQueryRequest], stream serverStream) error {
	if req.Msg.TransactionId != nil && *req.Msg.TransactionId != "" {
		s.lock.RLock()
		wrapper, ok := s.transactions[req.Msg.GetTransactionId()]
		s.lock.RUnlock()
		if !ok {
			return errors.WithStack(connect.NewError(connect.CodeNotFound, errors.Errorf("transaction %s not found", req.Msg.GetTransactionId())))
		}
		return errors.WithStack(s.executeQuery(ctx, wrapper.tx, req.Msg, stream))
	}
	return errors.WithStack(s.executeQuery(ctx, s.db, req.Msg, stream))
}

func (s *queryConn) executeQuery(ctx context.Context, db DB, req *querypb.ExecuteQueryRequest, stream serverStream) error {
	rawSQL, params, err := getSQLAndParams(req)
	if err != nil {
		return errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Wrap(err, "failed to parse parameters")))
	}

	switch req.CommandType {
	case querypb.CommandType_COMMAND_TYPE_EXEC:
		result, err := db.ExecContext(ctx, rawSQL, params...)
		if err != nil {
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to execute query")))
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to get rows affected")))
		}

		protoResp := &querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_ExecResult{
				ExecResult: &querypb.ExecResult{
					RowsAffected: rowsAffected,
				},
			},
		}

		err = stream.Send(protoResp)
		if err != nil {
			return errors.Wrap(err, "failed to send exec result")
		}

	case querypb.CommandType_COMMAND_TYPE_EXECRESULT:
		result, err := db.ExecContext(ctx, rawSQL, params...)
		if err != nil {
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to execute query")))
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to get rows affected")))
		}

		protoResp := &querypb.ExecuteQueryResponse{
			Result: &querypb.ExecuteQueryResponse_ExecResult{
				ExecResult: &querypb.ExecResult{
					RowsAffected: rowsAffected,
				},
			},
		}

		err = stream.Send(protoResp)
		if err != nil {
			return errors.Wrap(err, "failed to send exec result")
		}

	case querypb.CommandType_COMMAND_TYPE_ONE:
		row := db.QueryRowContext(ctx, rawSQL, params...)
		jsonRows, err := scanRowToMap(row, req.ResultColumns)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.WithStack(handleNoRows(stream))
			}
			return errors.Wrap(err, "failed to scan row")
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
			return errors.Wrap(err, "failed to send row results")
		}

	case querypb.CommandType_COMMAND_TYPE_MANY:
		rows, err := db.QueryContext(ctx, rawSQL, params...)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.WithStack(handleNoRows(stream))
			}
			return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to execute query")))
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
				return errors.WithStack(connect.NewError(connect.CodeInternal, errors.Wrap(err, "failed to scan row")))
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
				return errors.Wrap(err, "failed to send row results")
			}

			if !hasMore {
				rowCount = 0
			}
		}
	case querypb.CommandType_COMMAND_TYPE_UNSPECIFIED:
		return errors.WithStack(connect.NewError(connect.CodeInvalidArgument, errors.Errorf("unknown command type: %v", req.CommandType)))
	}
	return nil
}

var paramRe = regexp.MustCompile(`\?|/\*SLICE:[^*]*\*/\s*\?`)

// getSQLAndParams returns the SQL and parameters for a query.
// It handles SLICE patterns, which are used via SQLC to pass arrays to the database.
func getSQLAndParams(req *querypb.ExecuteQueryRequest) (string, []any, error) {
	params, err := parseJSONParameters(req.GetParametersJson())
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to parse parameters")
	}

	sql := req.RawSql

	// If no SLICE patterns, we don't need to do anything.
	if !strings.Contains(sql, "/*SLICE:") {
		return sql, params, nil
	}

	// Track which parameter index we're at as we scan the query
	paramIdx := 0
	var newParams []any

	// Replace each /*SLICE:xxx*/? with the right number of placeholders
	newSQL := paramRe.ReplaceAllStringFunc(sql, func(match string) string {
		if paramIdx >= len(params) {
			return match // Keep original if out of params
		}

		// Get the parameter that corresponds to this ?
		param := params[paramIdx]
		paramIdx++

		if match == "?" {
			newParams = append(newParams, param)
			return "?"
		}

		sliceVal := reflect.ValueOf(param)
		if sliceVal.Kind() != reflect.Slice && sliceVal.Kind() != reflect.Array {
			// Not a slice, keep original ? and add param as-is
			newParams = append(newParams, param)
			return "?"
		}

		sliceLen := sliceVal.Len()
		if sliceLen == 0 {
			return "NULL" // Empty slice case
		}

		// Add each slice element to our params
		for i := range sliceLen {
			newParams = append(newParams, sliceVal.Index(i).Interface())
		}

		// Generate ?, ?, ... with the right number of placeholders
		placeholders := strings.TrimSuffix(strings.Repeat("?, ", sliceLen), ", ")
		return placeholders
	})
	return newSQL, newParams, nil
}

func parseJSONParameters(paramsJSON string) ([]any, error) {
	if paramsJSON == "" {
		return nil, nil
	}

	var params []any
	if err := encoding.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal parameters")
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
	return time.Time{}, errors.Errorf("could not parse time string: %s", s)
}

// scanRowToMap scans a row and returns a JSON string representation
func scanRowToMap(row any, resultColumns []*querypb.ResultColumn) (string, error) {
	if len(resultColumns) == 0 {
		var rawValue any

		var err error
		switch r := row.(type) {
		case *sql.Row:
			err = r.Scan(&rawValue)
		case *sql.Rows:
			var columns []string
			columns, err = r.Columns()
			if err != nil {
				return "", errors.Wrap(err, "failed to get column names")
			}
			if len(columns) != 1 {
				return "", errors.Errorf("expected exactly one column for raw value query, got %d", len(columns))
			}
			err = r.Scan(&rawValue)
		default:
			return "", errors.Errorf("unsupported row type: %T", row)
		}
		if err != nil {
			return "", errors.Wrap(err, "failed to scan raw value")
		}

		if rawValue == nil {
			return "", nil
		}

		jsonBytes, err := encoding.Marshal(processFieldValue(rawValue))
		if err != nil {
			return "", errors.Wrap(err, "failed to marshal raw value")
		}
		return string(jsonBytes), nil
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
			return "", errors.Wrap(err, "failed to get column names")
		}
	case *sql.Row:
		// For sql.Row we can't get column names, but we know they must match our query
		dbColumns = sqlColumns
	default:
		return "", errors.Errorf("unsupported row type: %T", row)
	}

	if len(dbColumns) != len(resultColumns) {
		return "", errors.Errorf("column count mismatch: got %d columns from DB but expected %d columns", len(dbColumns), len(resultColumns))
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
		return "", errors.Wrap(err, "failed to scan row")
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

		field := structValue.FieldByName(exportName(typeName))
		if field.IsValid() {
			field.Set(reflect.ValueOf(processFieldValue(val)))
		}
	}

	jsonBytes, err := encoding.Marshal(structValue.Interface())
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal result")
	}

	return string(jsonBytes), nil
}

func processFieldValue(val any) any {
	switch v := val.(type) {
	case []byte:
		str := string(v)
		if t, err := parseTimeString(str); err == nil {
			return t
		}
		return str
	case string:
		if t, err := parseTimeString(v); err == nil {
			return t
		}
		return v
	case uint:
		// Convert safely to avoid overflow
		if v <= uint(math.MaxInt) {
			return int(v)
		}
		return float64(v)
	case uint8:
		return int(v) // Safe: uint8 max (255) fits in int
	case uint16:
		return int(v) // Safe: uint16 max (65535) fits in int
	case uint32:
		// On 32-bit platforms int is 32 bits, on 64-bit it's 64 bits
		if strconv.IntSize < 64 && v > uint32(math.MaxInt32) {
			return float64(v)
		}
		return int(v)
	case uint64:
		// Convert safely to avoid overflow
		if v <= uint64(math.MaxInt) {
			return int(v)
		}
		return float64(v)
	default:
		return v
	}
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
		return errors.Wrap(err, "failed to send no rows response")
	}
	return nil
}
