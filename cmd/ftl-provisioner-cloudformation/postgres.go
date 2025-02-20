package main

import (
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/block/ftl/internal/provisioner/state"
)

const PostgresPort = 5432

type PostgresTemplater struct {
	input  state.InputPostgres
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
		Port:                            ptr(PostgresPort),
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
