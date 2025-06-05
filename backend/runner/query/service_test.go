package query

import (
	"context"
	"database/sql"
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"
	_ "modernc.org/sqlite"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/schema"
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
			Verb:          &schema.Verb{Name: "getTestTypes", Response: &schema.Array{Element: &schema.Int{}}},
			RawSQL:        `SELECT id FROM test_table WHERE name IN (/*SLICE:names*/?) AND value = ? AND id IN (/*SLICE:ids*/?)`,
			CommandType:   Many,
			Parameters:    []any{[]string{"test1", "test3"}, 100, []int{1}},
			ResultColumns: []ResultColumn{{TypeName: "INT", SQLName: "id"}},
		})
		assert.NoError(t, err)
		strRows := []string{}
		for _, row := range jsonRows {
			strRows = append(strRows, string(row))
		}
		assert.Equal(t,
			[]string{
				"{\"int\":1}",
			},
			strRows,
		)
	})

	t.Run("TransactionLifecycle", func(t *testing.T) {
		txnKey, err := svc.BeginTransaction(ctx, "test_db")
		assert.NoError(t, err)
		assert.NotZero(t, txnKey)

		svc.lock.RLock()
		_, exists := svc.transactions[txnKey.String()]
		svc.lock.RUnlock()
		assert.True(t, exists)

		err = svc.CommitTransaction(ctx, txnKey)
		assert.NoError(t, err)

		svc.lock.RLock()
		_, exists = svc.transactions[txnKey.String()]
		svc.lock.RUnlock()
		assert.False(t, exists)
	})

	t.Run("TransactionRollback", func(t *testing.T) {
		txnKey, err := svc.BeginTransaction(ctx, "test_db")
		assert.NoError(t, err)
		assert.NotZero(t, txnKey)

		svc.lock.RLock()
		_, exists := svc.transactions[txnKey.String()]
		svc.lock.RUnlock()
		assert.True(t, exists)

		err = svc.RollbackTransaction(ctx, txnKey)
		assert.NoError(t, err)

		svc.lock.RLock()
		_, exists = svc.transactions[txnKey.String()]
		svc.lock.RUnlock()
		assert.False(t, exists)
	})

	t.Run("InvalidTransactionID", func(t *testing.T) {
		err := svc.CommitTransaction(ctx, key.NewTransactionKey("test_db", "invalid"))
		assert.Error(t, err)

		err = svc.RollbackTransaction(ctx, key.NewTransactionKey("test_db", "invalid"))
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
