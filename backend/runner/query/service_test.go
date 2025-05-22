package query

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	_ "modernc.org/sqlite"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT,
			value INTEGER
		)
	`)
	assert.NoError(t, err)

	// Insert some test data
	_, err = db.Exec(`
		INSERT INTO test_table (id, name, value) VALUES
		(1, 'test1', 100),
		(2, 'test2', 200),
		(3, 'test3', 300)
	`)
	assert.NoError(t, err)

	return db
}

func TestQueryService(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	t.Cleanup(func() {
		db.Close()
	})

	svc := &queryConn{
		transactions: make(map[string]*txWrapper),
		db:           db,
		lock:         sync.RWMutex{},
	}

	ctx := context.Background()

	t.Run("QuerySlice", func(t *testing.T) {
		jsonRows, _, err := svc.ExecuteQuery(ctx, ExecuteQueryRequest{
			RawSQL:         `SELECT id FROM test_table WHERE name IN (/*SLICE:names*/?) AND value = ? AND id IN (/*SLICE:ids*/?)`,
			CommandType:    Many,
			ParametersJSON: `[["test1", "test3"], 100, [1]]`,
			ResultColumns:  []ResultColumn{{TypeName: "INT", SQLName: "id"}},
		})
		assert.NoError(t, err)
		assert.Equal(t,
			[]string{
				"{\"int\":1}",
			},
			jsonRows,
		)
	})

	t.Run("TransactionLifecycle", func(t *testing.T) {
		beginResp, err := svc.BeginTransaction(ctx, connect.NewRequest(&querypb.BeginTransactionRequest{}))
		assert.NoError(t, err)
		assert.NotZero(t, beginResp.Msg.TransactionId)
		assert.Equal(t, querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS, beginResp.Msg.Status)

		txID := beginResp.Msg.TransactionId

		svc.lock.RLock()
		_, exists := svc.transactions[txID]
		svc.lock.RUnlock()
		assert.True(t, exists)

		commitResp, err := svc.CommitTransaction(ctx, connect.NewRequest(&querypb.CommitTransactionRequest{
			TransactionId: txID,
		}))
		assert.NoError(t, err)
		assert.Equal(t, querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS, commitResp.Msg.Status)

		svc.lock.RLock()
		_, exists = svc.transactions[txID]
		svc.lock.RUnlock()
		assert.False(t, exists)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		beginResp, err := svc.BeginTransaction(ctx, connect.NewRequest(&querypb.BeginTransactionRequest{}))
		assert.NoError(t, err)
		txID := beginResp.Msg.TransactionId

		svc.lock.RLock()
		_, exists := svc.transactions[txID]
		svc.lock.RUnlock()
		assert.True(t, exists)

		rollbackResp, err := svc.RollbackTransaction(ctx, connect.NewRequest(&querypb.RollbackTransactionRequest{
			TransactionId: txID,
		}))
		assert.NoError(t, err)
		assert.Equal(t, querypb.TransactionStatus_TRANSACTION_STATUS_SUCCESS, rollbackResp.Msg.Status)

		svc.lock.RLock()
		_, exists = svc.transactions[txID]
		svc.lock.RUnlock()
		assert.False(t, exists)
	})

	t.Run("InvalidTransactionID", func(t *testing.T) {
		_, err := svc.CommitTransaction(ctx, connect.NewRequest(&querypb.CommitTransactionRequest{
			TransactionId: "invalid",
		}))
		assert.Error(t, err)

		_, err = svc.RollbackTransaction(ctx, connect.NewRequest(&querypb.RollbackTransactionRequest{
			TransactionId: "invalid",
		}))
		assert.Error(t, err)
	})
}

func TestServiceConfig(t *testing.T) {
	t.Run("InvalidEngine", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = newQueryConn(t.Context(), "user:pass@tcp(localhost:3306)/db", "invalid") //nolint:errcheck
		}, "unsupported database engine: invalid")
	})

	t.Run("MissingDSN", func(t *testing.T) {
		_, err := newQueryConn(t.Context(), "", "mysql")
		assert.Error(t, err)
	})
}

type responseCollector struct {
	responses []*querypb.ExecuteQueryResponse
}

func (r *responseCollector) Send(resp *querypb.ExecuteQueryResponse) error {
	r.responses = append(r.responses, resp)
	return nil
}
