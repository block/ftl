package server

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/alecthomas/types/once"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/go-runtime/server/query"
	"github.com/block/ftl/internal/deploymentcontext"
	"github.com/block/ftl/internal/log"
)

func DatabaseHandle[T any](name, dbtype string) reflection.VerbResource {
	return func() reflect.Value {
		reflectedDB := reflection.GetDatabase[T]()
		db := ftl.NewDatabaseHandle[T](name, ftl.DatabaseType(dbtype), reflectedDB.DB)
		return reflect.ValueOf(db)
	}
}

func InitPostgres(ref reflection.Ref) (reflection.ReflectedDatabase, *once.Handle[*sql.DB]) {
	return InitDatabase(ref, "postgres", deploymentcontext.DBTypePostgres, "pgx")
}
func InitMySQL(ref reflection.Ref) (reflection.ReflectedDatabase, *once.Handle[*sql.DB]) {
	return InitDatabase(ref, "mysql", deploymentcontext.DBTypeMySQL, "mysql")
}

func InitDatabase(ref reflection.Ref, dbtype string, protoDBtype deploymentcontext.DBType, driver string) (reflection.ReflectedDatabase, *once.Handle[*sql.DB]) {
	return reflection.ReflectedDatabase{
			Name:   ref.Name,
			DBType: dbtype,
		},
		once.Once(func(ctx context.Context) (*sql.DB, error) {
			logger := log.FromContext(ctx)
			provider := deploymentcontext.FromContext(ctx).CurrentContext()
			dsn, testDB, err := provider.GetDatabase(ref.Name, protoDBtype)
			if err != nil {
				return nil, fmt.Errorf("failed to get database %q: %w", ref.Name, err)
			}

			logger.Debugf("Opening database: %s", ref.Name)
			db, err := otelsql.Open(driver, dsn)
			if err != nil {
				return nil, fmt.Errorf("failed to open database %q: %w", ref.Name, err)
			}

			// sets db.system and db.name attributes
			metricAttrs := otelsql.WithAttributes(
				semconv.DBSystemKey.String(dbtype),
				semconv.DBNameKey.String(ref.Name),
				attribute.Bool("ftl.is_user_service", true),
			)
			err = otelsql.RegisterDBStatsMetrics(db, metricAttrs)
			if err != nil {
				return nil, fmt.Errorf("failed to register database metrics: %w", err)
			}
			db.SetConnMaxIdleTime(time.Minute)
			if testDB {
				// In tests we always close the connections, as the DB being clean might invalidate pooled connections
				db.SetMaxIdleConns(0)
			} else {
				db.SetMaxOpenConns(20)
			}
			return db, nil
		})
}

// maybeBeginTransaction begins a transaction if the provided ref is a registered transaction verb
func maybeBeginTransaction(ctx context.Context, ref reflection.Ref) (string, error) {
	db, ok := reflection.GetTransactionDatabase(ref).Get()
	if !ok {
		return "", nil
	}
	txID, err := query.BeginTransaction(ctx, db.Name)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	return txID, nil
}

// maybeRollbackTransaction rolls back the current transaction if one is open
func maybeRollbackTransaction(ctx context.Context, ref reflection.Ref) error {
	db, ok := reflection.GetTransactionDatabase(ref).Get()
	if !ok {
		return nil
	}
	err := query.RollbackCurrentTransaction(ctx, db.Name)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}
	return nil
}

// maybeCommitTransaction commits the current transaction if one is open
func maybeCommitTransaction(ctx context.Context, ref reflection.Ref) error {
	db, ok := reflection.GetTransactionDatabase(ref).Get()
	if !ok {
		return nil
	}
	err := query.CommitCurrentTransaction(ctx, db.Name)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
