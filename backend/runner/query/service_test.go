package query

import (
	"context"
	"database/sql"
	"net/url"
	"sync"
	"testing"

	"connectrpc.com/connect"
	"github.com/alecthomas/assert/v2"
	_ "github.com/mattn/go-sqlite3"

	querypb "github.com/block/ftl/backend/protos/xyz/block/ftl/query/v1"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
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

	svc := &Service{
		transactions: make(map[string]*sql.Tx),
		db:           db,
		lock:         sync.RWMutex{},
	}

	ctx := context.Background()

	t.Run("TransactionLifecycle", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Run("ValidConfig", func(t *testing.T) {
		config := &Config{
			Endpoint: &url.URL{Scheme: "http", Host: "localhost:8896"},
			Engine:   "mysql",
			DSN:      "user:pass@tcp(localhost:3306)/db",
		}
		assert.NoError(t, config.Validate())
	})

	t.Run("InvalidEngine", func(t *testing.T) {
		config := &Config{
			Endpoint: &url.URL{Scheme: "http", Host: "localhost:8896"},
			Engine:   "invalid",
			DSN:      "user:pass@tcp(localhost:3306)/db",
		}
		assert.Error(t, config.Validate())
	})

	t.Run("MissingBind", func(t *testing.T) {
		config := &Config{
			Engine: "mysql",
			DSN:    "user:pass@tcp(localhost:3306)/db",
		}
		assert.Error(t, config.Validate())
	})

	t.Run("MissingDSN", func(t *testing.T) {
		config := &Config{
			Endpoint: &url.URL{Scheme: "http", Host: "localhost:8896"},
			Engine:   "mysql",
		}
		assert.Error(t, config.Validate())
	})
}
