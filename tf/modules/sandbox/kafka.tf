
# A single shared MSK cluster for the sandbox environment
resource "aws_msk_cluster" "ftl_msk_sandbox_cluster" {
  count = var.kafka_enabled ? 1 : 0
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
    security_groups = [aws_security_group.msk_sg[0].id]
    storage_info {
      ebs_storage_info {
        volume_size = 10
      }
    }
  }
}

resource "aws_security_group" "msk_sg" {
  count = var.kafka_enabled ? 1 : 0
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
