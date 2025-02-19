package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/block/ftl/cmd/ftl-provisioner-cloudformation/executor"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
)

type PostgresTemplater struct {
	input  executor.PostgresInputState
	config *Config
}

var _ ResourceTemplater = (*PostgresTemplater)(nil)

func (p *PostgresTemplater) AddToTemplate(template *goformation.Template) error {
	clusterID := cloudformationResourceID(p.input.ResourceID, "cluster")
	instanceID := cloudformationResourceID(p.input.ResourceID, "instance")
	template.Resources[clusterID] = &rds.DBCluster{
		Engine:                          ptr("aurora-postgresql"),
		MasterUsername:                  ptr("root"),
		ManageMasterUserPassword:        ptr(true),
		DBSubnetGroupName:               ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:             []string{p.config.DatabaseSecurityGroup},
		EngineMode:                      ptr("provisioned"),
		Port:                            ptr(5432),
		EnableIAMDatabaseAuthentication: ptr(true),
		ServerlessV2ScalingConfiguration: &rds.DBCluster_ServerlessV2ScalingConfiguration{
			MinCapacity: ptr(0.5),
			MaxCapacity: ptr(10.0),
		},
		Tags: ftlTags(p.input.Cluster, p.input.Module),
	}
	template.Resources[instanceID] = &rds.DBInstance{
		Engine:              ptr("aurora-postgresql"),
		DBInstanceClass:     ptr("db.serverless"),
		DBClusterIdentifier: ptr(goformation.Ref(clusterID)),
		Tags:                ftlTags(p.input.Cluster, p.input.Module),
	}
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "Endpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		PropertyName: PropertyPsqlWriteEndpoint,
		ResourceKind: ResourceKindPostgres,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		PropertyName: PropertyPsqlReadEndpoint,
		ResourceKind: ResourceKindPostgres,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		PropertyName: PropertyPsqlMasterUserARN,
		ResourceKind: ResourceKindPostgres,
	})
	return nil
}

type PostgresSetupExecutor struct {
	secrets *secretsmanager.Client
	inputs  []executor.State
}

func NewPostgresSetupExecutor(secrets *secretsmanager.Client) *PostgresSetupExecutor {
	return &PostgresSetupExecutor{secrets: secrets}
}

var _ executor.Executor = &PostgresSetupExecutor{}

func (e *PostgresSetupExecutor) Prepare(_ context.Context, input executor.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}

func (e *PostgresSetupExecutor) Execute(ctx context.Context) ([]executor.State, error) {
	rg := concurrency.ResourceGroup[executor.State]{}

	for _, input := range e.inputs {
		if input, ok := input.(executor.PostgresInstanceReadyState); ok {
			rg.Go(func() (executor.State, error) {
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

				return executor.PostgresDBDoneState{
					PostgresInstanceReadyState: input,
					Connector:                  connector,
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

	return endpointToDSN(&connector.Endpoint, connector.Database, adminUsername, adminPassword), nil
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
