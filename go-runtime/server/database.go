package server

import (
	"context"
	"database/sql"
	"reflect"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/once"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/ftl"
	"github.com/block/ftl/internal/deploymentcontext"
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
				return nil, errors.Wrapf(err, "failed to get database %q", ref.Name)
			}

			logger.Debugf("Opening database: %s", ref.Name)
			db, err := otelsql.Open(driver, dsn)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to open database %q", ref.Name)
			}

			// sets db.system and db.name attributes
			metricAttrs := otelsql.WithAttributes(
				semconv.DBSystemKey.String(dbtype),
				semconv.DBNameKey.String(ref.Name),
				attribute.Bool("ftl.is_user_service", true),
			)
			err = otelsql.RegisterDBStatsMetrics(db, metricAttrs)
			if err != nil {
				return nil, errors.Wrap(err, "failed to register database metrics")
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
