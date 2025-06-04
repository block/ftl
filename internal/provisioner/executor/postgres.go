package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/block/ftl/common/log"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

// ARNSecretPostgresSetup is an executor that sets up a postgres database on an RDS instance.
// It uses the secret ARN to get the admin username and password for the database.
type ARNSecretPostgresSetup struct {
	secrets  *secretsmanager.Client
	inputs   []state.State
	username string
}

func NewPostgresSetup(secrets *secretsmanager.Client, username string) *ARNSecretPostgresSetup {
	return &ARNSecretPostgresSetup{secrets: secrets, username: username}
}

var _ provisioner.Executor = &ARNSecretPostgresSetup{}

func (e *ARNSecretPostgresSetup) Prepare(_ context.Context, input state.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}

func (e *ARNSecretPostgresSetup) Execute(ctx context.Context) ([]state.State, error) {
	rg := concurrency.ResourceGroup[state.State]{}
	for _, input := range e.inputs {
		if input, ok := input.(state.RDSInstanceReadyPostgres); ok {
			rg.Go(func() (state.State, error) {
				connector := &schema.AWSIAMAuthDatabaseConnector{
					Database: input.ResourceID,
					Endpoint: input.WriteEndpoint,
					Username: e.username,
				}

				rootDSN, adminDSN, err := postgresAdminDSN(ctx, e.secrets, input.MasterUserSecretARN, connector)
				if err != nil {
					return nil, errors.WithStack(err)
				}

				// we need a separate connection to create the initial database
				if err := postgresEnsureDB(ctx, rootDSN, connector); err != nil {
					return nil, errors.WithStack(err)
				}

				// user credentials need to be set up after connecting to the new database
				if err := postgresSetup(ctx, adminDSN, connector); err != nil {
					return nil, errors.WithStack(err)
				}

				return state.OutputPostgres{
					Module:     input.Module,
					ResourceID: input.ResourceID,

					Connector: connector,
				}, nil
			})
		}
	}

	res, err := rg.Wait()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute postgres setup")
	}

	return res, nil
}

func postgresAdminDSN(ctx context.Context, secrets *secretsmanager.Client, secretARN string, connector *schema.AWSIAMAuthDatabaseConnector) (string, string, error) {
	adminUsername, adminPassword, err := secretARNToUsernamePassword(ctx, secrets, secretARN)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get username and password from secret ARN")
	}

	host, port, err := net.SplitHostPort(connector.Endpoint)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to split host and port")
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to convert port to int")
	}

	rootDSN := dsn.PostgresDSN("postgres", dsn.Host(host), dsn.Port(portInt), dsn.Username(adminUsername), dsn.Password(adminPassword))
	adminDSN := dsn.PostgresDSN(connector.Database, dsn.Host(host), dsn.Port(portInt), dsn.Username(adminUsername), dsn.Password(adminPassword))

	return rootDSN, adminDSN, nil
}

func postgresSetup(ctx context.Context, adminDSN string, connector *schema.AWSIAMAuthDatabaseConnector) error {
	logger := log.FromContext(ctx)

	database := connector.Database
	username := connector.Username

	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return errors.Wrap(err, "failed to connect to postgres")
	}
	defer db.Close()

	logger.Debugf("ensuring user %s exists", username)
	res, err := db.ExecContext(ctx, "CREATE USER "+username+" WITH LOGIN; GRANT rds_iam TO "+username+";")
	if err != nil {
		// Ignore if user already exists
		if !strings.Contains(err.Error(), "already exists") {
			return errors.Wrap(err, "failed to create database")
		}
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}
	if rows > 0 {
		logger.Debugf("postgres user created: %s", username)
	} else {
		logger.Debugf("postgres user already exists: %s", username)
	}

	logger.Debugf("granting privileges to user %s", username)
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				GRANT CONNECT ON DATABASE %s TO %s;
				GRANT USAGE ON SCHEMA public TO %s;
				GRANT CREATE ON SCHEMA public TO %s;
				GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s;
				GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %s;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;
			`, database, username, username, username, username, username, username, username)); err != nil {
		return errors.Wrap(err, "failed to grant FTL user privileges")
	}
	return nil
}

func postgresEnsureDB(ctx context.Context, rootDSN string, connector *schema.AWSIAMAuthDatabaseConnector) error {
	logger := log.FromContext(ctx).Scope("postgres-setup")
	database := connector.Database

	db, err := sql.Open("pgx", rootDSN)
	if err != nil {
		return errors.Wrap(err, "failed to connect to postgres")
	}
	defer db.Close()

	logger.Debugf("ensuring database %s exists", database)
	// Create the database if it doesn't exist
	res, err := db.ExecContext(ctx, "CREATE DATABASE "+database)
	if err != nil {
		// Ignore if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return errors.Wrap(err, "failed to create database")
		}
	} else {
		rows, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "failed to get rows affected")
		}
		if rows > 0 {
			logger.Infof("Database created: %s", database) //nolint:forbidigo
		} else {
			logger.Debugf("Database already exists: %s", database)
		}
	}
	return nil
}

func secretARNToUsernamePassword(ctx context.Context, secrets *secretsmanager.Client, secretARN string) (string, string, error) {
	secret, err := secrets.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get secret value")
	}
	secretString := *secret.SecretString

	var secretData map[string]string
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		return "", "", errors.Wrap(err, "failed to unmarshal secret data")
	}

	return secretData["username"], secretData["password"], nil
}
