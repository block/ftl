package provisioner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

var redPandaBrokers = []string{"127.0.0.1:19092"}
var pubSubNameLimit = 249 // 255 (filename limit) - 6 (partition id)

// NewDevProvisioner creates a new provisioner that provisions resources locally when running FTL in dev mode
func NewDevProvisioner(postgresPort int, mysqlPort int, recreate bool) *InMemProvisioner {
	// return NewEmbeddedProvisioner(map[schema.ResourceType]LegacyInMemResourceProvisionerFn{
	// 	schema.ResourceTypeTopic:        provisionTopic(),
	// 	schema.ResourceTypeSubscription: provisionSubscription(),
	// }, map[schema.ResourceType]InMemResourceProvisionerFn{
	// 	schema.ResourceTypePostgres: provisionPostgres(postgresPort, recreate),
	// 	schema.ResourceTypeMysql:    provisionMysql(mysqlPort, recreate),
	// })
	return NewEmbeddedProvisioner(map[schema.ResourceType]InMemResourceProvisionerFn{
		schema.ResourceTypePostgres: provisionPostgres(postgresPort, recreate),
		schema.ResourceTypeMysql:    provisionMysql(mysqlPort, recreate),
	})
}
func provisionMysql(mysqlPort int, recreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, res schema.Provisioned, module *schema.Module) (*RuntimeEvent, error) {
		logger := log.FromContext(ctx)

		dbName := strcase.ToLowerSnake(module.Name) + "_" + strcase.ToLowerSnake(res.ResourceID())

		logger.Infof("Provisioning mysql database: %s", dbName)

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
				return &RuntimeEvent{Database: &schema.DatabaseRuntimeEvent{
					ID:      res.ResourceID(),
					Payload: event,
				}}, nil
			}
		}
	}
}

func establishMySQLDB(ctx context.Context, mysqlDSN string, dbName string, mysqlPort int, recreate bool) (*schema.DatabaseRuntimeConnectionsEvent, error) {
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

	return &schema.DatabaseRuntimeConnectionsEvent{
		Connections: &schema.DatabaseRuntimeConnections{
			Write: &schema.DSNDatabaseConnector{DSN: dsn},
			Read:  &schema.DSNDatabaseConnector{DSN: dsn},
		},
	}, nil
}

func ProvisionPostgresForTest(ctx context.Context, module *schema.Module, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionPostgres(15432, true)(ctx, node, module)
	if err != nil {
		return "", err
	}

	return event.Database.Payload.(*schema.DatabaseRuntimeConnectionsEvent).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert
}

func ProvisionMySQLForTest(ctx context.Context, module *schema.Module, id string) (string, error) {
	node := &schema.Database{Name: id + "_test"}
	event, err := provisionMysql(13306, true)(ctx, node, module)
	if err != nil {
		return "", err
	}
	return event.Database.Payload.(*schema.DatabaseRuntimeConnectionsEvent).Connections.Write.(*schema.DSNDatabaseConnector).DSN, nil //nolint:forcetypeassert

}

func provisionPostgres(postgresPort int, recreate bool) InMemResourceProvisionerFn {
	return func(ctx context.Context, resource schema.Provisioned, module *schema.Module) (*RuntimeEvent, error) {
		logger := log.FromContext(ctx)

		dbName := strcase.ToLowerSnake(module.Name) + "_" + strcase.ToLowerSnake(resource.ResourceID())
		logger.Infof("Provisioning postgres database: %s", dbName)

		// We assume that the DB has already been started when running in dev mode
		postgresDSN := dsn.PostgresDSN("ftl", dsn.Port(postgresPort))

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
		return &RuntimeEvent{
			Database: &schema.DatabaseRuntimeEvent{
				ID: resource.ResourceID(),
				Payload: &schema.DatabaseRuntimeConnectionsEvent{
					Connections: &schema.DatabaseRuntimeConnections{
						Write: &schema.DSNDatabaseConnector{DSN: dsn},
						Read:  &schema.DSNDatabaseConnector{DSN: dsn},
					},
				},
			},
		}, nil
	}

}

// func provisionTopic() func(ctx context.Context, rc *provisioner.ResourceContext, module string, id string, previous *provisioner.Resource) (*provisioner.Resource, error) {
// 	return func(ctx context.Context, rc *provisioner.ResourceContext, module, id string, previous *provisioner.Resource) (*provisioner.Resource, error) {
// 		logger := log.FromContext(ctx)
// 		if err := dev.SetUpRedPanda(ctx); err != nil {
// 			return nil, fmt.Errorf("could not set up redpanda: %w", err)
// 		}
// 		topic, ok := rc.Resource.Resource.(*provisioner.Resource_Topic)
// 		if !ok {
// 			panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
// 		}

// 		topicID := kafkaTopicID(module, id)
// 		logger.Infof("Provisioning topic: %s", topicID)

// 		config := sarama.NewConfig()
// 		admin, err := sarama.NewClusterAdmin(redPandaBrokers, config)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to create cluster admin: %w", err)
// 		}
// 		defer admin.Close()

// 		topicMetas, err := admin.DescribeTopics([]string{topicID})
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to describe topic: %w", err)
// 		}
// 		if len(topicMetas) != 1 {
// 			return nil, fmt.Errorf("expected topic metadata from kafka but received none")
// 		}
// 		if topicMetas[0].Err == sarama.ErrUnknownTopicOrPartition {
// 			// No topic exists yet. Create it
// 			err = admin.CreateTopic(topicID, &sarama.TopicDetail{
// 				NumPartitions:     8,
// 				ReplicationFactor: 1,
// 				ReplicaAssignment: nil,
// 			}, false)
// 			if err != nil {
// 				return nil, fmt.Errorf("failed to create topic: %w", err)
// 			}
// 		} else if topicMetas[0].Err != sarama.ErrNoError {
// 			return nil, fmt.Errorf("failed to describe topic %q: %w", topicID, topicMetas[0].Err)
// 		}

// 		if topic.Topic == nil {
// 			topic.Topic = &provisioner.TopicResource{}
// 		}
// 		topic.Topic.Output = &provisioner.TopicResource_TopicResourceOutput{
// 			KafkaBrokers: redPandaBrokers,
// 			TopicId:      topicID,
// 		}
// 		return rc.Resource, nil
// 	}
// }

// func provisionSubscription() func(ctx context.Context, rc *provisioner.ResourceContext, module, id string, previous *provisioner.Resource) (*provisioner.Resource, error) {
// 	return func(ctx context.Context, rc *provisioner.ResourceContext, module, id string, previous *provisioner.Resource) (*provisioner.Resource, error) {
// 		logger := log.FromContext(ctx)
// 		if err := dev.SetUpRedPanda(ctx); err != nil {
// 			return nil, fmt.Errorf("could not set up redpanda: %w", err)
// 		}
// 		subscription, ok := rc.Resource.Resource.(*provisioner.Resource_Subscription)
// 		if !ok {
// 			panic(fmt.Errorf("unexpected resource type: %T", rc.Resource.Resource))
// 		}

// 		topicID := kafkaTopicID(subscription.Subscription.Topic.Module, subscription.Subscription.Topic.Name)
// 		consumerGroupID := consumerGroupID(module, id)
// 		subscription.Subscription.Output = &provisioner.SubscriptionResource_SubscriptionResourceOutput{
// 			KafkaBrokers:    redPandaBrokers,
// 			TopicId:         topicID,
// 			ConsumerGroupId: consumerGroupID,
// 		}
// 		logger.Infof("Provisioning subscription: %v", consumerGroupID)
// 		return rc.Resource, nil
// 	}
// }

// func kafkaTopicID(module, id string) string {
// 	return shortenString(fmt.Sprintf("%s.%s", module, id), pubSubNameLimit)
// }

// func consumerGroupID(module, id string) string {
// 	return shortenString(fmt.Sprintf("%s.%s", module, id), pubSubNameLimit)
// }

// shortenString truncates the input string to maxLength and appends a hash of the original string for uniqueness
func shortenString(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	hash := sha256.Sum256([]byte(input))
	hashStr := hex.EncodeToString(hash[:])
	truncateLength := maxLength - len(hashStr) - 1
	if truncateLength <= 0 {
		return hashStr[:maxLength]
	}
	return input[:truncateLength] + "-" + hashStr
}
