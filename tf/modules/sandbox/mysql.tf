# Used to allow runners to connect to provisioned MySQL instances
resource "aws_security_group" "mysql_security_group" {
  name        = "${var.ftl_cluster_name}-aurora-mysql-sg"
  description = "Allow MySQL access from EKS cluster"
  vpc_id = data.aws_vpc.selected.id

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected.cidr_block] # Restrict to VPC
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "mysql_security_group" {
  value = aws_security_group.mysql_security_group
}

output "aurora_security_group" {
  value = aws_security_group.postgres_security_group
}

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