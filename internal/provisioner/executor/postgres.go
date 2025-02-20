package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

// ARNSecretPostgresSetup is an executor that sets up a postgres database on an RDS instance.
// It uses the secret ARN to get the admin username and password for the database.
type ARNSecretPostgresSetup struct {
	secrets *secretsmanager.Client
	inputs  []state.State
}

func NewPostgresSetup(secrets *secretsmanager.Client) *ARNSecretPostgresSetup {
	return &ARNSecretPostgresSetup{secrets: secrets}
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
					Username: "ftluser",
				}

				adminDSN, err := postgresAdminDSN(ctx, e.secrets, input.MasterUserSecretARN, connector)
				if err != nil {
					return nil, err
				}

				if err := postgresSetup(ctx, adminDSN, connector); err != nil {
					return nil, err
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
		return nil, fmt.Errorf("failed to execute postgres setup: %w", err)
	}

	return res, nil
}

func postgresAdminDSN(ctx context.Context, secrets *secretsmanager.Client, secretARN string, connector *schema.AWSIAMAuthDatabaseConnector) (string, error) {
	adminUsername, adminPassword, err := secretARNToUsernamePassword(ctx, secrets, secretARN)
	if err != nil {
		return "", fmt.Errorf("failed to get username and password from secret ARN: %w", err)
	}

	host, port, err := net.SplitHostPort(connector.Endpoint)
	if err != nil {
		return "", fmt.Errorf("failed to split host and port: %w", err)
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return "", fmt.Errorf("failed to convert port to int: %w", err)
	}

	return dsn.PostgresDSN(connector.Database, dsn.Host(host), dsn.Port(portInt), dsn.Username(adminUsername), dsn.Password(adminPassword)), nil
}

func postgresSetup(ctx context.Context, adminDSN string, connector *schema.AWSIAMAuthDatabaseConnector) error {
	database := connector.Database
	username := connector.Username

	db, err := sql.Open("pgx", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	// Create the database if it doesn't exist
	if _, err := db.ExecContext(ctx, "CREATE DATABASE "+database); err != nil {
		// Ignore if database already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	if _, err := db.ExecContext(ctx, "CREATE USER "+username+" WITH LOGIN; GRANT rds_iam TO "+username+";"); err != nil {
		// Ignore if user already exists
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
				GRANT CONNECT ON DATABASE %s TO %s;
				GRANT USAGE ON SCHEMA public TO %s;
				GRANT CREATE ON SCHEMA public TO %s;
				GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO %s;
				GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO %s;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO %s;
				ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO %s;
			`, database, username, username, username, username, username, username, username)); err != nil {
		return fmt.Errorf("failed to grant FTL user privileges: %w", err)
	}
	return nil
}

func secretARNToUsernamePassword(ctx context.Context, secrets *secretsmanager.Client, secretARN string) (string, string, error) {
	secret, err := secrets.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretARN,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret value: %w", err)
	}
	secretString := *secret.SecretString

	var secretData map[string]string
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		return "", "", fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	return secretData["username"], secretData["password"], nil
}
