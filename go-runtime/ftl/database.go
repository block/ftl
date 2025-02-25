package ftl

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alecthomas/types/once"
	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	_ "github.com/jackc/pgx/v5/stdlib" // Register Postgres driver
)

type DatabaseType string

const (
	DatabaseTypePostgres DatabaseType = "postgres"
	DatabaseTypeMysql    DatabaseType = "mysql"
)

type DatabaseHandle[T any] struct {
	name  string
	_type DatabaseType
	db    *once.Handle[*sql.DB]
}

// Name returns the name of the database.
func (d DatabaseHandle[T]) Name() string { return d.name }

// Type returns the type of the database, e.g. "postgres"
func (d DatabaseHandle[T]) Type() DatabaseType {
	return d._type
}

// String returns a string representation of the database handle.
func (d DatabaseHandle[T]) String() string {
	return fmt.Sprintf("database %q", d.name)
}

// Get returns the SQL DB connection for the database.
func (d DatabaseHandle[T]) Get(ctx context.Context) *sql.DB {
	db, err := d.db.Get(ctx)
	if err != nil {
		panic(err)
	}
	return db
}

// NewDatabaseHandle is managed by FTL.
func NewDatabaseHandle[T any](name string, dbType DatabaseType, db *once.Handle[*sql.DB]) DatabaseHandle[T] {
	return DatabaseHandle[T]{name: name, db: db, _type: dbType}
}
