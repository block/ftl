terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
    }
  }
}

data "aws_vpc" "selected" {
  filter {
    name   = "tag:Name"
    values = [var.vpc_name]
  }
}

data "aws_subnets" "selected" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.selected.id]
  }
  filter {
    name   = "tag:kubernetes.io/role/internal-elb"
    values = ["1"]
  }
}

output "values" {
  value = yamlencode(
    {
      provisioner: {
        config: {
          plugins: [
            {
              id : "cloudformation"
              resources : [
                "postgres"
              ],
            },{
              id: "sandbox"
              resources: concat(
                ["mysql"],
                var.kafka_enabled ? ["topic", "subscription"] : []
              )
            }]
        }
        env: concat([
            {
              name: "FTL_PROVISIONER_PLUGIN_CONFIG_FILE",
              value: "/config/config.toml"
            },
            {
              name: "FTL_PROVISIONER_CF_DB_SUBNET_GROUP",
              value: aws_db_subnet_group.aurora-postgres-subnet-group.name
            },
            {
              name: "FTL_PROVISIONER_CF_MYSQL_SECURITY_GROUP",
              value: aws_security_group.mysql_security_group.id
            },
            {
              name: "FTL_PROVISIONER_CF_DB_SECURITY_GROUP",
              value: aws_security_group.postgres_security_group.id
            },
            {
              name: "FTL_SANDBOX_MYSQL_ARN",
              value: aws_rds_cluster.ftl_mysql_sandbox_db.master_user_secret[0].secret_arn
            },
            {
              name: "FTL_SANDBOX_MYSQL_ENDPOINT",
              value: "${aws_rds_cluster_instance.ftl_mysql_sandbox_instance[0].endpoint}:${aws_rds_cluster_instance.ftl_mysql_sandbox_instance[0].port}"
            },
            {
              name: "TMPDIR",
              value: "/working"
            }
          ],
          var.kafka_enabled ? [{
            name: "FTL_SANDBOX_KAFKA_BROKERS",
            value: aws_msk_cluster.ftl_msk_sandbox_cluster[0].bootstrap_brokers
          }] : []
        )
      }
    }
  )
}