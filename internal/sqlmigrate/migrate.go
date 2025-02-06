// Package sqlmigrate supports a dbmate-compatible superset of migration files.
//
// The superset is that in addition to a migration being a .sql file, it can
// also be a Go function which is called to execute the migration.
package sqlmigrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/block/ftl/internal/log"
)

var migrationFileNameRe = regexp.MustCompile(`^.*(\d{14})_(.*)\.sql$`)

type Migration func(ctx context.Context, db *sql.Tx) error

type namedMigration struct {
	name      string
	version   string
	migration string
}

func (m namedMigration) String() string { return m.name }

// Migrate applies all migrations in the provided fs.FS to the provided database.
func Migrate(ctx context.Context, dsn *url.URL, migrationFiles fs.FS) error {
	dsnString := dsn.String()
	switch dsn.Scheme {
	case "sqlite", "mysql":
		dsnString = strings.TrimPrefix(dsnString, dsn.Scheme+"://")
	}
	db, err := sql.Open(dsn.Scheme, dsnString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close() //nolint:errcheck

	// Create schema_migrations table if it doesn't exist.
	// This table structure is compatible with dbmate.
	_, _ = db.ExecContext(ctx, `CREATE TABLE schema_migrations (version TEXT PRIMARY KEY)`) //nolint:errcheck

	sqlFiles, err := fs.Glob(migrationFiles, "*.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	migrations := make([]namedMigration, 0, len(sqlFiles))

	// Collect .sql files.
	for _, sqlFile := range sqlFiles {
		name := filepath.Base(sqlFile)
		groups := migrationFileNameRe.FindStringSubmatch(name)
		if groups == nil {
			return fmt.Errorf("invalid migration file name %q, must be in the form <date>_<detail>.sql", sqlFile)
		}
		version := groups[1]
		sqlMigration, err := fs.ReadFile(migrationFiles, sqlFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %q: %w", sqlFile, err)
		}
		migrations = append(migrations, namedMigration{name, version, string(sqlMigration)})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	for _, migration := range migrations {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migration %s: failed to begin transaction: %w", migration, err)
		}
		err = applyMigration(ctx, dsn, tx, migration)
		if err != nil {
			if txerr := tx.Rollback(); txerr != nil {
				return fmt.Errorf("migration %s: failed to rollback transaction: %w", migration, txerr)
			}
			return fmt.Errorf("migration %s: %w", migration, err)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("migration %s: failed to commit transaction: %w", migration, err)
		}
	}
	return nil
}

func applyMigration(ctx context.Context, dsn *url.URL, tx *sql.Tx, migration namedMigration) error {
	start := time.Now()
	logger := log.FromContext(ctx).Scope("migrate:" + migration.name)
	ctx = log.ContextWithLogger(ctx, logger)
	var err error
	switch dsn.Scheme {
	case "postgres":
		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", migration.version)
	case "mysql":
		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON DUPLICATE KEY UPDATE version=version", migration.version)
	case "sqlite":
		_, err = tx.ExecContext(ctx, "INSERT OR IGNORE INTO schema_migrations (version) VALUES ($1)", migration.version)
	default:
		panic(fmt.Sprintf("unsupported database driver %T", dsn.Scheme))
	}
	if err != nil {
		return fmt.Errorf("failed to insert migration: %w", err)
	}
	logger.Debugf("Applying migration")
	if err := migrateSQLFile(ctx, tx, migration.name, migration.migration); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	logger.Debugf("Applied migration in %s", time.Since(start))
	return nil
}

func migrateSQLFile(ctx context.Context, db *sql.Tx, name string, sqlMigration string) error {
	logger := log.FromContext(ctx)
	logger.Debugf("%s", strings.ReplaceAll(strings.TrimSpace(sqlMigration), "\n", "\\n"))
	_, err := db.ExecContext(ctx, sqlMigration)
	if err != nil {
		return fmt.Errorf("failed to execute migration %q: %w", name, err)
	}
	return nil
}
