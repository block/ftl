package executor

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/go-sql-driver/mysql"

	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
	"github.com/block/ftl/internal/dsn"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/provisioner"
	"github.com/block/ftl/internal/provisioner/state"
)

const passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_+=<>?"
const usernameChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ARNSecretMySQLSetup is an executor that sets up a mysql database on an RDS instance.
// It uses the secret ARN to get the admin username and password for the database.
type ARNSecretMySQLSetup struct {
	secrets  *secretsmanager.Client
	inputs   []state.State
	username string
}

func NewARNSecretMySQLSetup(secrets *secretsmanager.Client, username string) *ARNSecretMySQLSetup {
	return &ARNSecretMySQLSetup{secrets: secrets, username: username}
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
				databaseName := createDatabaseName(input.Module, input.ResourceID)
				connector := &schema.AWSIAMAuthDatabaseConnector{
					Database: databaseName,
					Endpoint: input.WriteEndpoint,
					Username: e.username,
				}
				if arn, ok := input.MasterUserSecretARN.Get(); ok {

					adminDSN, err := mysqlAdminDSN(ctx, e.secrets, arn, connector)
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
				}
				dsn, err := mysqlDSNSetup(ctx, input.WriteEndpoint, databaseName)
				if err != nil {
					return nil, err
				}
				return state.OutputMySQL{
					Module:     input.Module,
					ResourceID: input.ResourceID,
					Connector: &schema.DSNDatabaseConnector{
						Database: databaseName,
						DSN:      dsn,
					},
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

func createDatabaseName(module string, resourceID string) string {
	sanitisedModule := strings.ReplaceAll(module, "_", "")
	sanitisedResourceID := strings.ReplaceAll(resourceID, "_", "")

	return fmt.Sprintf("%s_%s", sanitisedModule, sanitisedResourceID)
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
	logger := log.FromContext(ctx)

	database := connector.Database
	username := connector.Username

	db, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql: %w", err)
	}
	defer db.Close()

	res, err := db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+database)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows > 0 {
		logger.Infof("MySQL database created: %s", database) //nolint:forbidigo
	} else {
		logger.Debugf("MySQL database already exists: %s", database)
	}

	res, err = db.ExecContext(ctx, "CREATE USER IF NOT EXISTS "+username+"@'%' IDENTIFIED WITH AWSAuthenticationPlugin AS 'RDS';")
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	rows, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows > 0 {
		logger.Debugf("MySQL user created: %s", username)
	} else {
		logger.Debugf("MySQL user already exists: %s", username)
	}

	if _, err := db.ExecContext(ctx, "ALTER USER "+username+"@'%' REQUIRE SSL;"); err != nil {
		return fmt.Errorf("failed to require ssl: %w", err)
	}

	if _, err := db.ExecContext(ctx, "USE "+database+";"); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "GRANT ALL PRIVILEGES ON "+database+".* TO "+username+"@'%';"); err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	return nil
}

func mysqlDSNSetup(ctx context.Context, adminDSN string, database string) (string, error) {
	parsed, err := mysql.ParseDSN(adminDSN)
	if err != nil {
		return "", fmt.Errorf("failed to parse dsn: %w", err)
	}

	db, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return "", fmt.Errorf("failed to connect to mysql: %w", err)
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+database)
	if err != nil {
		return "", fmt.Errorf("failed to create database: %w", err)
	}
	pw, err := generatePassword(16, passwordChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}
	username, err := generatePassword(8, usernameChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate username: %w", err)
	}
	_, err = db.ExecContext(ctx, "CREATE USER IF NOT EXISTS "+username+"@'%' IDENTIFIED BY '"+pw+"';")
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	if _, err := db.ExecContext(ctx, "USE "+database+";"); err != nil {
		return "", fmt.Errorf("failed to use database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "GRANT ALL PRIVILEGES ON "+database+".* TO "+username+"@'%';"); err != nil {
		return "", fmt.Errorf("failed to grant privileges: %w", err)
	}
	parsed.User = username
	parsed.Passwd = pw
	parsed.DBName = database
	return parsed.FormatDSN(), nil
}

func generatePassword(length int, alphabet string) (string, error) {
	password := make([]byte, length)
	charCount := big.NewInt(int64(len(alphabet)))

	for i := range length {
		index, err := rand.Int(rand.Reader, charCount)
		if err != nil {
			return "", fmt.Errorf("failed to generate password: %w", err)
		}
		password[i] = alphabet[index.Int64()]
	}

	return string(password), nil
}
