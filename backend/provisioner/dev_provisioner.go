package provisioner

import (
	"context"
	"database/sql"
	"time"

	"github.com/XSAM/otelsql"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/types/optional"
	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl/common/key"
	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/provisioner/executor"
	"github.com/block/ftl/internal/provisioner/state"
)

var redPandaBrokers = []string{"127.0.0.1:19092"}

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int, mysqlPort int, recreate bool) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypePostgres:     provisionPostgres(postgresPort, recreate),
		schema.ResourceTypeMysql:        provisionMysql(mysqlPort, recreate),
		schema.ResourceTypeTopic:        provisionTopic(),
		schema.ResourceTypeSubscription: provisionSubscription(),
		schema.ResourceTypeImage:        provisionDummyImage(),
	}, map[schema.ResourceType]InMemResourceProvisionerFn{})
}

func provisionDummyImage() InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		return &schema.RuntimeElement{Element: &schema.ModuleRuntimeImage{Image: "n/a"}, Deployment: deployment}, nil
	}
}
func provisionMysql(mysqlPort int, recreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx).Deployment(deployment)

		dbName := strcase.ToLowerSnake(deployment.Payload.Module) + "_" + strcase.ToLowerSnake(res.ResourceID())

		logger.Infof("Provisioning mysql database: %s", dbName) //nolint
		db, ok := res.(*schema.Database)
		if !ok {
			return nil, errors.Errorf("expected database, got %T", res)
		}
		migrationHash := ""
		for migration := range slices.FilterVariants[*schema.MetadataSQLMigration](db.Metadata) {
			migrationHash = migration.Digest
			break
		}

		// We assume that the DB hsas already been started when running in dev mode
		mysqlDSN, err := dev.SetupMySQL(ctx, mysqlPort)
		if err != nil {
			return nil, errors.Wrap(err, "failed to wait for mysql to be ready")
		}
		timeout := time.After(10 * time.Second)
		retry := time.NewTicker(100 * time.Millisecond)
		defer retry.Stop()
		for {
			select {
			case <-timeout:
				return nil, errors.Wrap(err, "failed to query database")
			case <-retry.C:
				event, err := establishMySQLDB(ctx, mysqlDSN, dbName, mysqlPort, recreate, migrationHash)
				if err != nil {
					logger.Debugf("failed to establish mysql database: %s", err.Error())
					continue
				}
				return &schema.RuntimeElement{
					Deployment: deployment,
					Name:       optional.Some(res.ResourceID()),
					Element: &schema.DatabaseRuntime{
						Connections: event,
					}}, nil
			}
		}
	}
}

func establishMySQLDB(ctx context.Context, mysqlDSN string, dbName string, mysqlPort int, recreate bool, migrationHash string) (*schema.DatabaseRuntimeConnections, error) {
	conn, err := otelsql.Open("mysql", mysqlDSN)
	logger := log.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to mysql")
	}
	defer conn.Close()

	res, err := conn.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query database")
	}
	defer res.Close()

	exists := res.Next()

	if migrationHash != "" {
		_, err := conn.Exec("CREATE TABLE IF NOT EXISTS migrations (db VARCHAR(255) PRIMARY KEY NOT NULL, migration VARCHAR(255) NOT NULL)")
		if err != nil {
			return nil, errors.Wrap(err, "failed to create migrations tracking table")
		}
		if exists && !recreate {
			// We might still need to recreate the database if the schema has changed
			existing := ""
			err := conn.QueryRow("SELECT migration FROM migrations WHERE db=?", dbName).Scan(&existing)
			if err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return nil, errors.Wrap(err, "failed to query migrations table")
				}
				logger.Debugf("No existing migration found")
			} else {
				logger.Debugf("existing migration: %s , current migration %s", existing, migrationHash)
				if existing != migrationHash {
					logger.Infof("Recreating database %q due to schema change", dbName) //nolint
					recreate = true
				}
			}
		}
	}

	if exists && recreate {
		_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to drop database %q", dbName)
		}
	}
	if !exists || recreate {
		_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create database %q", dbName)
		}
	}

	dsn := dsn.MySQLDSN(dbName, dsn.Port(mysqlPort))

	if migrationHash != "" {
		_, err := conn.Exec("INSERT INTO migrations (db, migration) VALUES (?, ?) ON DUPLICATE KEY UPDATE migration = ?", dbName, migrationHash, migrationHash)
		if err != nil {
			return nil, errors.Wrap(err, "failed to insert migration hash")
		}
	}

	return &schema.DatabaseRuntimeConnections{
		Write: &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
		Read:  &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
	}, nil
}

func ProvisionPostgresForTest(ctx context.Context, realm, module, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionPostgres(15432, true)(ctx, key.NewChangesetKey(), key.NewDeploymentKey(realm, module), node, nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return event.Element.(*schema.DatabaseRuntime).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert
}

func ProvisionMySQLForTest(ctx context.Context, realm, module, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionMysql(13306, true)(ctx, key.NewChangesetKey(), key.NewDeploymentKey(realm, module), node, nil)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return event.Element.(*schema.DatabaseRuntime).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert

}

func provisionPostgres(postgresPort int, alwaysRecreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, resource schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		recreate := alwaysRecreate
		logger := log.FromContext(ctx).Deployment(deployment)
		db, ok := resource.(*schema.Database)
		if !ok {
			return nil, errors.Errorf("expected database, got %T", resource)
		}

		dbName := strcase.ToLowerSnake(deployment.Payload.Module) + "_" + strcase.ToLowerSnake(resource.ResourceID())
		logger.Infof("Provisioning postgres database: %s", dbName) //nolint

		// We assume that the DB has already been started when running in dev mode
		postgresDSN := dsn.PostgresDSN("ftl", dsn.Port(postgresPort))
		err := dev.SetupPostgres(ctx, optional.None[string](), postgresPort, recreate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to wait for postgres to be ready")
		}

		conn, err := otelsql.Open("pgx", postgresDSN)
		if err != nil {
			return nil, errors.Wrap(err, "failed to connect to postgres")
		}
		defer conn.Close()

		res, err := conn.Query("SELECT * FROM pg_catalog.pg_database WHERE datname=$1", dbName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to query database")
		}
		defer res.Close()

		exists := res.Next()
		migrationHash := ""
		for migration := range slices.FilterVariants[*schema.MetadataSQLMigration](db.Metadata) {
			_, err := conn.Exec("CREATE TABLE IF NOT EXISTS migrations (db VARCHAR PRIMARY KEY NOT NULL, migration VARCHAR NOT NULL)")
			if err != nil {
				return nil, errors.Wrap(err, "failed to create migrations tracking table")
			}
			migrationHash = migration.Digest
			if exists && !recreate {
				// We might still need to recreate the database if the schema has changed
				existing := ""
				err := conn.QueryRow("SELECT migration FROM migrations WHERE db=$1", dbName).Scan(&existing)
				if err != nil {
					if !errors.Is(err, sql.ErrNoRows) {
						return nil, errors.Wrap(err, "failed to query migrations table")
					}
				} else {
					if existing != migrationHash {
						logger.Infof("Recreating database %q due to schema change", dbName) //nolint
						recreate = true
					}
				}
			}
		}

		if exists && recreate {
			// Terminate any dangling connections.
			_, err = conn.ExecContext(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
				dbName)
			if err != nil {
				return nil, errors.Wrap(err, "failed to kill existing backends")
			}
			_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to drop database %q", dbName)
			}
		}
		if !exists || recreate {
			_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create database %q", dbName)
			}
		}

		if migrationHash != "" {
			_, err := conn.Exec("INSERT INTO migrations (db, migration) VALUES ($1, $2)ON CONFLICT (db) DO UPDATE SET migration = EXCLUDED.migration;", dbName, migrationHash)
			if err != nil {
				return nil, errors.Wrap(err, "failed to insert migration hash")
			}
		}
		dsn := dsn.PostgresDSN(dbName, dsn.Port(postgresPort))

		return &schema.RuntimeElement{
			Name:       optional.Some(resource.ResourceID()),
			Deployment: deployment,
			Element: &schema.DatabaseRuntime{
				Connections: &schema.DatabaseRuntimeConnections{
					Write: &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
					Read:  &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
				},
			}}, nil
	}

}

func provisionTopic() InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		if err := dev.SetUpRedPanda(ctx); err != nil {
			return nil, errors.Wrap(err, "could not set up redpanda")
		}

		topic, ok := res.(*schema.Topic)
		if !ok {
			return nil, errors.Errorf("expected topic, got %T", res)
		}

		partitions := 1
		if pmd, ok := slices.FindVariant[*schema.MetadataPartitions](topic.Metadata); ok {
			partitions = pmd.Partitions
		}

		exec := executor.NewKafkaTopicSetup()
		err := exec.Prepare(ctx, state.TopicClusterReady{
			InputTopic: state.InputTopic{
				Topic:      res.ResourceID(),
				Module:     deployment.Payload.Module,
				Partitions: partitions,
			},
			Brokers: redPandaBrokers,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to prepare kafka topic setup")
		}
		output, err := exec.Execute(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to execute kafka topic setup")
		}
		if len(output) != 1 {
			return nil, errors.Errorf("expected 1 output but got %d", len(output))
		}
		outputTopic, ok := output[0].(state.OutputTopic)
		if !ok {
			return nil, errors.Errorf("expected output topic but got %T", output[0])
		}
		return &schema.RuntimeElement{
			Name:       optional.Some(res.ResourceID()),
			Deployment: deployment,
			Element:    outputTopic.Runtime,
		}, nil
	}
}

func provisionSubscription() InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned, module *schema.Module) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)
		verb, ok := res.(*schema.Verb)
		if !ok {
			panic(errors.Errorf("unexpected resource type: %T", res))
		}
		for range slices.FilterVariants[*schema.MetadataSubscriber](verb.Metadata) {
			logger.Debugf("Provisioning subscription for verb: %s", verb.Name)
			return &schema.RuntimeElement{
				Name:       optional.Some(res.ResourceID()),
				Deployment: deployment,
				Element: &schema.PlaintextKafkaSubscriptionConnector{
					KafkaBrokers: redPandaBrokers,
				}}, nil
		}
		return nil, nil
	}
}
