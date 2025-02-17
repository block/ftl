package main

import (
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/rds"
	"github.com/block/ftl/cmd/ftl-provisioner-cloudformation/executor"
	_ "github.com/go-sql-driver/mysql"
)

type MySQLTemplater struct {
	input  executor.MySQLInputState
	config *Config
}

var _ ResourceTemplater = (*MySQLTemplater)(nil)

func (p *MySQLTemplater) AddToTemplate(template *goformation.Template) error {
	clusterID := cloudformationResourceID(p.input.ResourceID, "cluster")
	instanceID := cloudformationResourceID(p.input.ResourceID, "instance")
	template.Resources[clusterID] = &rds.DBCluster{
		Engine:                   ptr("aurora-mysql"),
		MasterUsername:           ptr("root"),
		ManageMasterUserPassword: ptr(true),
		DBSubnetGroupName:        ptr(p.config.DatabaseSubnetGroupARN),
		VpcSecurityGroupIds:      []string{p.config.DatabaseSecurityGroup},
		EngineMode:               ptr("provisioned"),
		Port:                     ptr(3306),
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
