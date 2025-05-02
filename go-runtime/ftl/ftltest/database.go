package ftltest

import (
	"context"
	"os"

	errors "github.com/alecthomas/errors"

	"github.com/block/ftl/backend/provisioner"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/internal/deploymentcontext"
)

// WithDatabase sets up a database for testing by appending "_test" to the DSN and emptying all tables
func WithDatabase[T any]() Option {
	return func(ctx context.Context, state *OptionsState) error {
		db := reflection.GetDatabase[T]()
		return errors.WithStack(setupTestDatabase(ctx, state, db.DBType, db.Name))
	}
}

// WithAllDatabases sets up all databases in the module for testing by appending "_test" to the DSN and emptying all tables
func WithAllDatabases() Option {
	return func(ctx context.Context, state *OptionsState) error {
		dbs := reflection.GetAllDatabases()
		for _, db := range dbs {
			if err := setupTestDatabase(ctx, state, db.DBType, db.Name); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}

func setupTestDatabase(ctx context.Context, state *OptionsState, dbType, dbName string) error {
	if _, ok := state.databases[dbName]; ok {
		return nil
	}
	switch dbType {
	case "postgres":
		dsn, err := provisioner.ProvisionPostgresForTest(ctx, state.project, moduleGetter(), "test", dbName)
		if err != nil {
			return errors.Wrapf(err, "could not provision database %q", dbName)
		}
		dir, err := os.Getwd()
		if err != nil {
			return errors.Errorf("could not get working dir")
		}
		err = provisioner.RunPostgresMigration(ctx, dsn, dir, dbName)
		if err != nil {
			return errors.Wrapf(err, "could not migrate database %q", dbName)
		}
		// replace original database with test database
		replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypePostgres, dsn)
		if err != nil {
			return errors.Wrapf(err, "could not create database %q with DSN %q", dbName, dsn)
		}
		state.databases[dbName] = replacementDB
	case "mysql":
		dsn, err := provisioner.ProvisionMySQLForTest(ctx, state.project, moduleGetter(), "test", dbName)
		if err != nil {
			return errors.Wrapf(err, "could not provision database %q", dbName)
		}
		dir, err := os.Getwd()
		if err != nil {
			return errors.Errorf("could not get working dir")
		}
		err = provisioner.RunMySQLMigration(ctx, dsn, dir, dbName)
		if err != nil {
			return errors.Wrapf(err, "could not migrate database %q", dbName)
		}
		// replace original database with test database
		replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypeMySQL, dsn)
		if err != nil {
			return errors.Wrapf(err, "could not create database %q with DSN %q", dbName, dsn)
		}
		state.databases[dbName] = replacementDB
	}
	return nil
}

// WithSQLVerbsEnabled enables direct execution of SQL verbs (without mocks). It also sets up any associated databases for tests.
func WithSQLVerbsEnabled() Option {
	return func(ctx context.Context, state *OptionsState) error {
		state.allowDirectSQLVerbs = true
		dbs := reflection.GetQueryVerbDatabases()
		for _, db := range dbs {
			if err := setupTestDatabase(ctx, state, db.DBType, db.Name); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	}
}
