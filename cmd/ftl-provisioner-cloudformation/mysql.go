package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
	_ "github.com/go-sql-driver/mysql"

	"github.com/block/ftl/cmd/ftl-provisioner-cloudformation/executor"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/concurrency"
	"github.com/block/ftl/internal/dsn"
)

const MySQLPort = 3306

type MySQLTemplater struct {
	input  executor.MySQLInputState
	config *Config
}

var _ ResourceTemplater = (*MySQLTemplater)(nil)

func (p *MySQLTemplater) AddToTemplate(template *goformation.Template) error {
	clusterID := cloudformationResourceID(p.input.ResourceID, "cluster")
	instanceID := cloudformationResourceID(p.input.ResourceID, "instance")
	template.Resources[clusterID] = &rds.DBCluster{
		Engine:                          ptr("aurora-mysql"),
		MasterUsername:                  ptr("root"),
		ManageMasterUserPassword:        ptr(true),
		DBSubnetGroupName:               ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:             []string{p.config.MysqlSecurityGroup},
		EngineMode:                      ptr("provisioned"),
		EnableIAMDatabaseAuthentication: ptr(true),
		Port:                            ptr(MySQLPort),
		ServerlessV2ScalingConfiguration: &rds.DBCluster_ServerlessV2ScalingConfiguration{
			MinCapacity: ptr(0.5),
			MaxCapacity: ptr(10.0),
		},
		Tags: ftlTags(p.input.Cluster, p.input.Module),
	}
	template.Resources[instanceID] = &rds.DBInstance{
		Engine:              ptr("aurora-mysql"),
		DBInstanceClass:     ptr("db.serverless"),
		DBClusterIdentifier: ptr(goformation.Ref(clusterID)),
		Tags:                ftlTags(p.input.Cluster, p.input.Module),
	}
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "Endpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		ResourceKind: ResourceKindMySQL,
		PropertyName: PropertyMySQLWriteEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "ReadEndpoint.Address"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		ResourceKind: ResourceKindMySQL,
		PropertyName: PropertyMySQLReadEndpoint,
	})
	addOutput(template.Outputs, goformation.GetAtt(clusterID, "MasterUserSecret.SecretArn"), &CloudformationOutputKey{
		ResourceID:   p.input.ResourceID,
		ResourceKind: ResourceKindMySQL,
		PropertyName: PropertyMySQLMasterUserARN,
	})
	return nil
}

type MySQLSetupExecutor struct {
	secrets *secretsmanager.Client
	inputs  []executor.State
}

func NewMySQLSetupExecutor(secrets *secretsmanager.Client) *MySQLSetupExecutor {
	return &MySQLSetupExecutor{secrets: secrets}
}

var _ executor.Executor = (*MySQLSetupExecutor)(nil)

func (e *MySQLSetupExecutor) Prepare(_ context.Context, input executor.State) error {
	e.inputs = append(e.inputs, input)
	return nil
}

func (e *MySQLSetupExecutor) Execute(ctx context.Context) ([]executor.State, error) {
	rg := concurrency.ResourceGroup[executor.State]{}

	for _, input := range e.inputs {
		if input, ok := input.(executor.MySQLInstanceReadyState); ok {
			rg.Go(func() (executor.State, error) {
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

				return executor.MySQLDBDoneState{
					MySQLInstanceReadyState: input,
					Connector:               connector,
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
