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
## Infrastructure for a sandbox environment
##
## This is an environment intended to be used for fast development and testing.
## It is not intended to be used for production.

# A single shared MySQL database for the sandbox environment
resource "aws_rds_cluster" "ftl_mysql_sandbox_db" {
  cluster_identifier                  = "${var.ftl_cluster_name}-mysql-sandbox-db"
  engine                              = "aurora-mysql"
  engine_version                      = "8.0.mysql_aurora.3.04.0"
  database_name                       = "ftl"
  manage_master_user_password         = true
  master_username                     = "root"
  iam_database_authentication_enabled = true
  backup_retention_period             = 7
  preferred_backup_window             = "07:00-09:00"
  db_subnet_group_name                = aws_db_subnet_group.aurora-postgres-subnet-group.name
  final_snapshot_identifier = "${var.ftl_cluster_name}-sandbox-msql-final-snapshot"
  # Add required settings for Aurora Serverless v2
  serverlessv2_scaling_configuration {
    min_capacity = 1
    max_capacity = 8
  }
  storage_encrypted    = true
  apply_immediately    = true
  vpc_security_group_ids = [aws_security_group.mysql_security_group.id]
  copy_tags_to_snapshot = true
  tags = {
    "ftl:cluster" = var.ftl_cluster_name
  }
}

resource "aws_rds_cluster_instance" "ftl_mysql_sandbox_instance" {
  count              = 1
  identifier         = "${var.ftl_cluster_name}-mysql-sandbox-instance-1"
  cluster_identifier = aws_rds_cluster.ftl_mysql_sandbox_db.id
  instance_class     = "db.serverless"
  engine             = aws_rds_cluster.ftl_mysql_sandbox_db.engine
  apply_immediately  = true
  copy_tags_to_snapshot = true
  
  tags = {
    "ftl:cluster" = var.ftl_cluster_name
  }
}

# A single shared MSK cluster for the sandbox environment
resource "aws_msk_cluster" "ftl_msk_sandbox_cluster" {
  cluster_name = "${var.ftl_cluster_name}-msk-sandbox-cluster"
  kafka_version = "3.5.1"
  number_of_broker_nodes = 3
  encryption_info {
    encryption_in_transit {
      # TODO: Change to TLS when we support that
      client_broker = "TLS_PLAINTEXT"
    }
  }
  broker_node_group_info {
    instance_type = "kafka.t3.small"
    client_subnets = data.aws_subnets.selected.ids
    security_groups = [aws_security_group.msk_sg.id]
    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

resource "aws_security_group" "msk_sg" {
  name        = "${var.ftl_cluster_name}-msk-sg"
  description = "Allow MSK access from EKS cluster"
  vpc_id = data.aws_vpc.selected.id

  ingress {
    from_port   = 9092
    to_port     = 9092
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected.cidr_block]  # Restrict to VPC
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
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
              resources: [
                "mysql",
                "topic",
                "subscription"
              ],
            }]
        }
        env: [{
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
            name: "FTL_SANDBOX_KAFKA_BROKERS",
            value: aws_msk_cluster.ftl_msk_sandbox_cluster.bootstrap_brokers
          },
          {
            name: "TMPDIR",
            value: "/working"
          }]
      }
    }
  )
}