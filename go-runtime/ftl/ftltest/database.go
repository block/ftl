package ftltest

import (
	"context"
	"fmt"
	"os"

	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/deploymentcontext"
)

// WithDatabase sets up a database for testing by appending "_test" to the DSN and emptying all tables
func WithDatabase[T any]() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			db := reflection.GetDatabase[T]()
			return setupTestDatabase(ctx, state, db.DBType, db.Name)
		},
	}
}

// WithAllDatabases sets up all databases in the module for testing by appending "_test" to the DSN and emptying all tables
func WithAllDatabases() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			dbs := reflection.GetAllDatabases()
			for _, db := range dbs {
				if err := setupTestDatabase(ctx, state, db.DBType, db.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func setupTestDatabase(ctx context.Context, state *OptionsState, dbType, dbName string) error {
	if _, ok := state.databases[dbName]; ok {
		return nil
	}
	switch dbType {
	case "postgres":
		dsn, err := provisioner.ProvisionPostgresForTest(ctx, moduleGetter(), dbName)
		if err != nil {
			return fmt.Errorf("could not provision database %q: %w", dbName, err)
		}
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get working dir")
		}
		err = provisioner.RunPostgresMigration(ctx, dsn, dir, dbName)
		if err != nil {
			return fmt.Errorf("could not migrate database %q: %w", dbName, err)
		}
		// replace original database with test database
		replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypePostgres, dsn)
		if err != nil {
			return fmt.Errorf("could not create database %q with DSN %q: %w", dbName, dsn, err)
		}
		state.databases[dbName] = replacementDB
	case "mysql":
		dsn, err := provisioner.ProvisionMySQLForTest(ctx, moduleGetter(), dbName)
		if err != nil {
			return fmt.Errorf("could not provision database %q: %w", dbName, err)
		}
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("could not get working dir")
		}
		err = provisioner.RunMySQLMigration(ctx, dsn, dir, dbName)
		if err != nil {
			return fmt.Errorf("could not migrate database %q: %w", dbName, err)
		}
		// replace original database with test database
		replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypeMySQL, dsn)
		if err != nil {
			return fmt.Errorf("could not create database %q with DSN %q: %w", dbName, dsn, err)
		}
		state.databases[dbName] = replacementDB
	}
	return nil
}

// WithSQLVerbsEnabled enables direct execution of SQL verbs (without mocks). It also sets up any associated databases for tests.
func WithSQLVerbsEnabled() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			state.allowDirectSQLVerbs = true
			dbs := reflection.GetQueryVerbDatabases()
			for _, db := range dbs {
				if err := setupTestDatabase(ctx, state, db.DBType, db.Name); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
