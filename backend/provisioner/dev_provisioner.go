package provisioner

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/XSAM/otelsql"
	"github.com/alecthomas/types/optional"
	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/common/slices"
	"github.com/block/ftl/common/strcase"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/key"
	"github.com/block/ftl/internal/log"
)

var redPandaBrokers = []string{"127.0.0.1:19092"}

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int, mysqlPort int, recreate bool) *InMemProvisioner {
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypePostgres:     provisionPostgres(postgresPort, recreate),
		schema.ResourceTypeMysql:        provisionMysql(mysqlPort, recreate),
		schema.ResourceTypeTopic:        provisionTopic(),
		schema.ResourceTypeSubscription: provisionSubscription(),
	}, map[schema.ResourceType]InMemResourceProvisionerFn{})
}
func provisionMysql(mysqlPort int, recreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)

		dbName := strcase.ToLowerSnake(deployment.Payload.Module) + "_" + strcase.ToLowerSnake(res.ResourceID())

		logger.Debugf("Provisioning mysql database: %s", dbName)

		// We assume that the DB hsas already been started when running in dev mode
		mysqlDSN, err := dev.SetupMySQL(ctx, mysqlPort)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for mysql to be ready: %w", err)
		}
		timeout := time.After(10 * time.Second)
		retry := time.NewTicker(100 * time.Millisecond)
		defer retry.Stop()
		for {
			select {
			case <-timeout:
				return nil, fmt.Errorf("failed to query database: %w", err)
			case <-retry.C:
				event, err := establishMySQLDB(ctx, mysqlDSN, dbName, mysqlPort, recreate)
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

func establishMySQLDB(ctx context.Context, mysqlDSN string, dbName string, mysqlPort int, recreate bool) (*schema.DatabaseRuntimeConnections, error) {
	conn, err := otelsql.Open("mysql", mysqlDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}
	defer conn.Close()

	res, err := conn.Query("SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?", dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer res.Close()

	exists := res.Next()
	if exists && recreate {
		_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to drop database %q: %w", dbName, err)
		}
	}
	if !exists || recreate {
		_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to create database %q: %w", dbName, err)
		}
	}

	dsn := dsn.MySQLDSN(dbName, dsn.Port(mysqlPort))

	return &schema.DatabaseRuntimeConnections{
		Write: &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
		Read:  &schema.DSNDatabaseConnector{DSN: dsn, Database: dbName},
	}, nil
}

func ProvisionPostgresForTest(ctx context.Context, moduleName string, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionPostgres(15432, true)(ctx, key.NewChangesetKey(), key.NewDeploymentKey(moduleName), node)
	if err != nil {
		return "", err
	}

	return event.Element.(*schema.DatabaseRuntime).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert
}

func ProvisionMySQLForTest(ctx context.Context, moduleName string, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionMysql(13306, true)(ctx, key.NewChangesetKey(), key.NewDeploymentKey(moduleName), node)
	if err != nil {
		return "", err
	}
	return event.Element.(*schema.DatabaseRuntime).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert

}

func provisionPostgres(postgresPort int, recreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, resource schema.Provisioned) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)

		dbName := strcase.ToLowerSnake(deployment.Payload.Module) + "_" + strcase.ToLowerSnake(resource.ResourceID())
		logger.Debugf("Provisioning postgres database: %s", dbName)

		// We assume that the DB has already been started when running in dev mode
		postgresDSN := dsn.PostgresDSN("ftl", dsn.Port(postgresPort))
		err := dev.SetupPostgres(ctx, optional.None[string](), postgresPort, recreate)
		if err != nil {
			return nil, fmt.Errorf("failed to wait for postgres to be ready: %w", err)
		}

		conn, err := otelsql.Open("pgx", postgresDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
		defer conn.Close()

		res, err := conn.Query("SELECT * FROM pg_catalog.pg_database WHERE datname=$1", dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to query database: %w", err)
		}
		defer res.Close()

		exists := res.Next()
		if exists && recreate {
			// Terminate any dangling connections.
			_, err = conn.ExecContext(ctx, `
			SELECT pid, pg_terminate_backend(pid)
			FROM pg_stat_activity
			WHERE datname = $1 AND pid <> pg_backend_pid()`,
				dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to kill existing backends: %w", err)
			}
			_, err = conn.ExecContext(ctx, "DROP DATABASE "+dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to drop database %q: %w", dbName, err)
			}
		}
		if !exists || recreate {
			_, err = conn.ExecContext(ctx, "CREATE DATABASE "+dbName)
			if err != nil {
				return nil, fmt.Errorf("failed to create database %q: %w", dbName, err)
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
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)
		if err := dev.SetUpRedPanda(ctx); err != nil {
			return nil, fmt.Errorf("could not set up redpanda: %w", err)
		}
		topic, ok := res.(*schema.Topic)
		if !ok {
			panic(fmt.Errorf("unexpected resource type: %T", res))
		}

		topicID := fmt.Sprintf("%s.%s", deployment.Payload.Module, topic.Name)
		logger.Debugf("Provisioning topic: %s", topicID)

		config := sarama.NewConfig()
		admin, err := sarama.NewClusterAdmin(redPandaBrokers, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka admin client: %w", err)
		}
		defer admin.Close()

		topicMetas, err := admin.DescribeTopics([]string{topicID})
		if err != nil {
			return nil, fmt.Errorf("failed to describe topic: %w", err)
		}
		if len(topicMetas) != 1 {
			return nil, fmt.Errorf("expected topic metadata from kafka but received none")
		}
		partitions := 1
		if pm, ok := slices.FindVariant[*schema.MetadataPartitions](topic.Metadata); ok {
			partitions = pm.Partitions
		}
		if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
			// No topic exists yet. Create it
			err = admin.CreateTopic(topicID, &sarama.TopicDetail{
				NumPartitions:     int32(partitions),
				ReplicationFactor: 1,
				ReplicaAssignment: nil,
			}, false)
			if err != nil {
				return nil, fmt.Errorf("failed to create topic: %w", err)
			}
		} else if topicMetas[0].Err != sarama.ErrNoError {
			return nil, fmt.Errorf("failed to describe topic %q: %w", topicID, topicMetas[0].Err)
		} else if len(topicMetas[0].Partitions) > partitions {
			var plural string
			if len(topicMetas[0].Partitions) == 1 {
				plural = "partition"
			} else {
				plural = "partitions"
			}
			logger.Warnf("Using existing topic %s with %d %s instead of %d", topicID, len(topicMetas[0].Partitions), plural, partitions)
		} else if len(topicMetas[0].Partitions) < partitions {
			if err := admin.CreatePartitions(topicID, int32(partitions), nil, false); err != nil {
				return nil, fmt.Errorf("failed to increase partitions: %w", err)
			}
		}

		return &schema.RuntimeElement{
			Name:       optional.Some(res.ResourceID()),
			Deployment: deployment,
			Element: &schema.TopicRuntime{
				KafkaBrokers: redPandaBrokers,
				TopicID:      topicID,
			}}, nil
	}
}

func provisionSubscription() InMemResourceProvisionerFn {
	return func(ctx context.Context, changeset key.Changeset, deployment key.Deployment, res schema.Provisioned) (*schema.RuntimeElement, error) {
		logger := log.FromContext(ctx)
		verb, ok := res.(*schema.Verb)
		if !ok {
			panic(fmt.Errorf("unexpected resource type: %T", res))
		}
		for range slices.FilterVariants[*schema.MetadataSubscriber](verb.Metadata) {
			logger.Debugf("Provisioning subscription for verb: %s", verb.Name)
			return &schema.RuntimeElement{
				Name:       optional.Some(res.ResourceID()),
				Deployment: deployment,
				Element: &schema.VerbRuntime{
					Subscription: &schema.VerbRuntimeSubscription{
						KafkaBrokers: redPandaBrokers,
					},
				}}, nil
		}
		return nil, nil
	}
}
