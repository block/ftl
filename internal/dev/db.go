package dev

import (
	"context"
	stdsql "database/sql"
	_ "embed"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	errors "github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/block/ftl/internal/container"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/observability"
)

//go:embed docker-compose.mysql.yml
var mysqlDockerCompose string

// use this lock while checking latestMysqlDSN status and running `docker compose up` if needed
var mysqlLock = &sync.Mutex{}
var latestMysqlDSN optional.Option[string]

//go:embed docker-compose.postgres.yml
var postgresDockerCompose string

// use this lock while checking latestPostgresDSN status and running `docker compose up` if needed
var postgresLock = &sync.Mutex{}
var latestPostgresDSN optional.Option[string]

// CreateForDevel creates and migrates a new database for development or testing.
//
// If "recreate" is true, the database will be dropped and recreated.
func CreateForDevel(ctx context.Context, dsn string, recreate bool) (*stdsql.DB, error) {
	logger := log.FromContext(ctx)
	config, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse DSN")
	}

	noDBDSN := *config
	noDBDSN.Path = "" // Remove the database name.

	var conn *stdsql.DB
	for range 10 {
		conn, err = observability.OpenDBAndInstrument(noDBDSN.String())
		if err == nil {
			defer conn.Close()
			break
		}
		logger.Debugf("Waiting for database to be ready: %v", err)
		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "ctx canclled")

		case <-time.After(1 * time.Second):
		}
	}
	if conn == nil {
		return nil, errors.Wrap(err, "database not ready after 10 tries")
	}

	dbName := strings.TrimPrefix(config.Path, "/")

	if recreate {
		// Terminate any dangling connections.
		_, err = conn.ExecContext(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
			dbName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to terminate connections")
		}

		_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %q", dbName))
		if err != nil {
			return nil, errors.Wrap(err, "failed to drop database")
		}
	}

	_, _ = conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %q", dbName)) //nolint:errcheck // PG doesn't support "IF NOT EXISTS" so instead we just ignore any error.

	realConn, err := observability.OpenDBAndInstrument(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}
	realConn.SetMaxIdleConns(20)
	realConn.SetMaxOpenConns(20)

	return realConn, nil
}

func PostgresDSN(ctx context.Context, port int) string {
	if port == 0 {
		port = 15432
	}
	return dsn.PostgresDSN("ftl", dsn.Port(port))
}

func SetupPostgres(ctx context.Context, image optional.Option[string], port int, recreate bool) error {
	postgresLock.Lock()
	defer postgresLock.Unlock()

	// check if already started with this port
	dsn := PostgresDSN(ctx, port)
	if latestDSN, ok := latestPostgresDSN.Get(); ok && latestDSN == dsn {
		return nil
	}
	envars := []string{}
	if port != 0 {
		envars = append(envars, "POSTGRES_PORT="+strconv.Itoa(port))
	}
	if imaneName, ok := image.Get(); ok {
		envars = append(envars, "FTL_DATABASE_IMAGE="+imaneName)
	}
	_, err := container.ComposeUp(ctx, "postgres", postgresDockerCompose, optional.None[string](), envars...)
	if err != nil {
		return errors.Wrap(err, "could not start postgres")
	}
	_, err = CreateForDevel(ctx, dsn, recreate)
	if err != nil {
		return errors.Wrap(err, "failed to create database")
	}
	latestPostgresDSN = optional.Some(dsn)
	return nil
}

func SetupMySQL(ctx context.Context, port int) (string, error) {
	mysqlLock.Lock()
	defer mysqlLock.Unlock()

	// check if already started with this port
	dsn := dsn.MySQLDSN("ftl", dsn.Port(port))
	if latestDSN, ok := latestMysqlDSN.Get(); ok && latestDSN == dsn {
		return dsn, nil
	}
	envars := []string{}
	if port != 0 {
		envars = append(envars, "MYSQL_PORT="+strconv.Itoa(port))
	}
	_, err := container.ComposeUp(ctx, "mysql", mysqlDockerCompose, optional.None[string](), envars...)
	if err != nil {
		return "", errors.Wrap(err, "could not start mysql")
	}

	log.FromContext(ctx).Debugf("MySQL DSN: %s", dsn)
	latestMysqlDSN = optional.Some(dsn)
	return dsn, nil
}
