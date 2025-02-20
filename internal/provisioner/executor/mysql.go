package executor

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

// ARNSecretMySQLSetup is an executor that sets up a mysql database on an RDS instance.
// It uses the secret ARN to get the admin username and password for the database.
type ARNSecretMySQLSetup struct {
	secrets *secretsmanager.Client
	inputs  []state.State
}

func NewARNSecretMySQLSetup(secrets *secretsmanager.Client) *ARNSecretMySQLSetup {
	return &ARNSecretMySQLSetup{secrets: secrets}
}

var _ provisioner.Executor = (*ARNSecretMySQLSetup)(nil)

func (e *ARNSecretMySQLSetup) Prepare(_ context.Context, input state.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}

func (e *ARNSecretMySQLSetup) Execute(ctx context.Context) ([]state.State, error) {
	rg := concurrency.ResourceGroup[state.State]{}

	for _, input := range e.inputs {
		if input, ok := input.(state.RDSInstanceReadyMySQL); ok {
			rg.Go(func() (state.State, error) {
				connector := &schema.AWSIAMAuthDatabaseConnector{
					Database: input.ResourceID,
					Endpoint: input.WriteEndpoint,
					Username: "ftluser",
				}

				adminDSN, err := mysqlAdminDSN(ctx, e.secrets, input.MasterUserSecretARN, connector)
				if err != nil {
					return nil, err
				}

				if err := mysqlSetup(ctx, adminDSN, connector); err != nil {
					return nil, err
				}

				return state.OutputMySQL{
					Module:     input.Module,
					ResourceID: input.ResourceID,

					Connector: connector,
				}, nil
			})
		}
	}

	res, err := rg.Wait()
	if err != nil {
		return nil, fmt.Errorf("failed to execute MySQL setup: %w", err)
	}

	return res, nil
}

func mysqlAdminDSN(ctx context.Context, secrets *secretsmanager.Client, secretARN string, connector *schema.AWSIAMAuthDatabaseConnector) (string, error) {
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

	return dsn.MySQLDSN("", dsn.Host(host), dsn.Port(portInt), dsn.Username(adminUsername), dsn.Password(adminPassword)), nil
}

func mysqlSetup(ctx context.Context, adminDSN string, connector *schema.AWSIAMAuthDatabaseConnector) error {
	database := connector.Database
	username := connector.Username

	db, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql: %w", err)
	}
	defer db.Close()

	if _, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+database); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "CREATE USER IF NOT EXISTS "+username+"@'%' IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';"); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if _, err := db.ExecContext(ctx, "ALTER USER "+username+"@'%' REQUIRE SSL;"); err != nil {
		return fmt.Errorf("failed to require ssl: %w", err)
	}

	if _, err := db.ExecContext(ctx, "USE "+database+";"); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "GRANT ALL PRIVILEGES ON "+database+" TO "+username+"@'%';"); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	return nil
}
