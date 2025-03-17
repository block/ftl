
resource "aws_db_subnet_group" "aurora-postgres-subnet-group" {
  name       = "${var.ftl_cluster_name}-aurora-postgres-subnet-group"
  subnet_ids = data.aws_subnets.selected.ids
}

resource "aws_security_group" "postgres_security_group" {
  name        = "${var.ftl_cluster_name}-aurora-postgres-sg"
  description = "Allow PostgreSQL access from EKS cluster"
  vpc_id = data.aws_vpc.selected.id

  ingress {
    from_port   = 5432
    to_port     = 5432
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

output "aurora-postgres-subnet-group" {
  value = aws_db_subnet_group.aurora-postgres-subnet-group
}

output "mysql_security_group" {
  value = aws_security_group.mysql_security_group
}

output "aurora_security_group" {
  value = aws_security_group.postgres_security_group
}