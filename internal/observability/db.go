package observability

import (
	"database/sql"

	"github.com/XSAM/otelsql"
	errors "github.com/alecthomas/errors"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func OpenDBAndInstrument(dsn string) (*sql.DB, error) {
	conn, err := otelsql.Open("pgx", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	err = otelsql.RegisterDBStatsMetrics(conn, otelsql.WithAttributes(semconv.DBSystemPostgreSQL))
	if err != nil {
		return nil, errors.Wrap(err, "failed to register database metrics")
	}

	return conn, nil
}
