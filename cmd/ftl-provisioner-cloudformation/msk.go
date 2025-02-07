package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	goformation "github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/ec2"
	"github.com/awslabs/goformation/v7/cloudformation/iam"
	"github.com/awslabs/goformation/v7/cloudformation/msk"
	"github.com/awslabs/goformation/v7/cloudformation/tags"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/key"
)

type MSKTopicTemplater struct {
	resourceID string
	cluster    string
	module     string
	config     *Config
}

var _ ResourceTemplater = (*MSKTopicTemplater)(nil)

func (m *MSKTopicTemplater) AddToTemplate(template *goformation.Template) error {
	mskClusterID, err := addMSKClusterToTemplate(template, m.cluster, m.config)
	if err != nil {
		return err
	}

	// TODO: Define IAM policies here

	addOutput(template.Outputs, goformation.GetAtt(mskClusterID, "Arn"), &CloudformationOutputKey{
		ResourceID:   m.resourceID,
		PropertyName: PropertyMSKCluserARN,
		ResourceKind: ResourceKindTopic,
	})

	return nil
}

func updateTopicOutputs(_ context.Context, deployment key.Deployment, resourceID string, outputs []types.Output) ([]*schema.RuntimeElement, error) {
	byName, err := outputsByPropertyName(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to group outputs by property name: %w", err)
	}

	brokers, err := getBootstrapBrokers(context.Background(), *byName[PropertyMSKCluserARN].OutputValue)
	if err != nil {
		return nil, err
	}
	event := schema.RuntimeElement{
		Deployment: deployment,
		Name:       optional.Some(resourceID),
		Element: &schema.TopicRuntime{
			KafkaBrokers: brokers,
		},
	}
	return []*schema.RuntimeElement{&event}, nil
}

type MSKSubscriptionTemplater struct {
	resourceID string
	cluster    string
	module     string
	config     *Config
}

var _ ResourceTemplater = (*MSKSubscriptionTemplater)(nil)

func (m *MSKSubscriptionTemplater) AddToTemplate(template *goformation.Template) error {
	mskClusterID, err := addMSKClusterToTemplate(template, m.cluster, m.config)
	if err != nil {
		return err
	}

	addOutput(template.Outputs, goformation.GetAtt(mskClusterID, "Arn"), &CloudformationOutputKey{
		ResourceID:   m.resourceID,
		PropertyName: PropertyMSKCluserARN,
		ResourceKind: ResourceKindSubscription,
	})

	return nil
}

func updateSubscriptionOutputs(_ context.Context, deployment key.Deployment, resourceID string, outputs []types.Output) ([]*schema.RuntimeElement, error) {
	byName, err := outputsByPropertyName(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to group outputs by property name: %w", err)
	}
	brokers, err := getBootstrapBrokers(context.Background(), *byName[PropertyMSKCluserARN].OutputValue)
	if err != nil {
		return nil, err
	}
	event := schema.RuntimeElement{
		Deployment: deployment,
		Name:       optional.Some(resourceID),
		Element: &schema.VerbRuntime{
			Subscription: &schema.VerbRuntimeSubscription{
				KafkaBrokers: brokers,
			},
		},
	}
	return []*schema.RuntimeElement{&event}, nil
}

func addMSKClusterToTemplate(template *goformation.Template, ftlCluster string, config *Config) (string, error) {
	mskClusterID := cloudformationResourceID("msk")
	seceurityGroupID := cloudformationResourceID("msk", "sg")
	iamPolicyID := cloudformationResourceID("msk", "provisioner", "policy")
	clusterTag := ftlClusterTag(ftlCluster)

	// create msk security group
	template.Resources[seceurityGroupID] = &ec2.SecurityGroup{
		GroupDescription: "Security group for FTL MSK",
		GroupName:        ptr("ftl-msk-sg"),
		SecurityGroupEgress: []ec2.SecurityGroup_Egress{
			{
				Description: ptr("Allow all outbound traffic"),
				CidrIp:      ptr("0.0.0.0/0"),
				FromPort:    ptr(0),
				ToPort:      ptr(0),
				IpProtocol:  "-1",
			},
		},
		SecurityGroupIngress: []ec2.SecurityGroup_Ingress{
			{
				Description: ptr("Restrict to VPC"),
				CidrIp:      ptr("0.0.0.0/16"),
				FromPort:    ptr(9092),
				ToPort:      ptr(9092),
				IpProtocol:  "tcp",
			},
		},
		Tags:  []tags.Tag{clusterTag},
		VpcId: ptr(config.VPCID),
	}

	// define IAM provisioner policy and attach it to the provisioner role
	template.Resources[iamPolicyID] = &iam.RolePolicy{
		PolicyDocument: map[string]interface{}{
			// TODO: Look at: https://docs.aws.amazon.com/msk/latest/developerguide/create-iam-role.html
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect": "Allow",
					"Action": []string{
						"kafka:CreateTopic",
						"kafka:DeleteTopic",
						"kafka:DescribeTopic",
						"kafka:ListTopics",
						"kafka:AlterTopic",
					},
					// TODO: what is the arn for the msk cluster?
					"Resource": "*",
				},
			},
		},
		PolicyName: "msk-provisioner-policy",
		// TODO: is this right??
		RoleName: goformation.Ref(cloudformationResourceID("provisioner", "role")),
	}

	// create msk cluster
	template.Resources[mskClusterID] = &msk.ServerlessCluster{
		ClusterName: "ftl-msk",
		ClientAuthentication: &msk.ServerlessCluster_ClientAuthentication{
			Sasl: &msk.ServerlessCluster_Sasl{
				Iam: &msk.ServerlessCluster_Iam{
					Enabled: true,
				},
			},
		},
		VpcConfigs: []msk.ServerlessCluster_VpcConfig{
			{
				SecurityGroups: []string{},
				SubnetIds:      config.MSKSubnetIDs,
			},
		},
		Tags: map[string]string{
			clusterTag.Key: clusterTag.Value,
		},
	}

	// TODO: allow admin to manage consumer groups?

	return mskClusterID, nil
}

func getBootstrapBrokers(ctx context.Context, clusterArn string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load default aws config: %w", err)
	}
	client := kafka.NewFromConfig(cfg)

	result, err := client.GetBootstrapBrokers(ctx, &kafka.GetBootstrapBrokersInput{
		ClusterArn: &clusterArn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bootstrap brokers: %w", err)
	}
	brokers := strings.Split(*result.BootstrapBrokerString, ",")
	return brokers, nil
}
