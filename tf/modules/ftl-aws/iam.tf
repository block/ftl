locals {
  controller_service_account_name = "${var.ftl_cluster_name}-controller"
  provisioner_service_account_name = "${var.ftl_cluster_name}-provisioner"
  admin_service_account_name = "${var.ftl_cluster_name}-admin"
}

data "aws_region" "current" {

}


data "aws_iam_policy_document" "assume_controller_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.kube_namespace}:${local.controller_service_account_name}"]
    }
  }
}

data "aws_iam_policy_document" "assume_admin_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.kube_namespace}:${local.admin_service_account_name}"]
    }
  }
}

resource "aws_iam_role" "ftl_admin_role" {
  name               = "${var.ftl_cluster_name}-admin-role"
  assume_role_policy = data.aws_iam_policy_document.assume_admin_role_policy.json
}

resource "aws_iam_role" "ftl_controller_role" {
  name               = "${var.ftl_cluster_name}-controller-role"
  assume_role_policy = data.aws_iam_policy_document.assume_controller_role_policy.json
}

resource "aws_iam_policy" "ftl_kms_policy" {
  name        = "${var.ftl_cluster_name}-kms-policy"
  description = "Policy to allow KMS key usage"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["sts:AssumeRole", "sts:AssumeRoleWithWebIdentity"]
        Effect   = "Allow"
        Resource = aws_iam_role.ftl_controller_role.arn
      },
      {
        Action   = [
          "secretsmanager:ListSecrets",
          "secretsmanager:CreateSecret",
          "secretsmanager:UpdateSecret",
          "secretsmanager:DeleteSecret",
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:PutSecretValue",
          "secretsmanager:BatchGetSecretValue",
          "secretsmanager:TagResource"
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action   = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:GenerateDataKey*",
          "kms:DescribeKey"
        ]
        Effect   = "Allow"
        Resource = aws_kms_key.ftl_db_kms_key.arn
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "attach_ftl_controller_kms_policy" {
  role       = aws_iam_role.ftl_controller_role.name
  policy_arn = aws_iam_policy.ftl_kms_policy.arn
}

resource "aws_iam_role_policy_attachment" "attach_ftl_admin_kms_policy" {
  role       = aws_iam_role.ftl_admin_role.name
  policy_arn = aws_iam_policy.ftl_kms_policy.arn
}

resource "aws_iam_role_policy" "ftl_controller_ecr_read_access" {
  role   = aws_iam_role.ftl_controller_role.name
  policy = data.aws_iam_policy_document.ecr_deployment_read_policy.json
}

resource "aws_iam_role_policy" "ftl_runner_ecr_read_access" {
  role   = aws_iam_role.ftl_runner_role.name
  policy = data.aws_iam_policy_document.ecr_deployment_read_policy.json
}

resource "aws_iam_role_policy" "ftl_provisioner_ecr_read_access" {
  role   = aws_iam_role.ftl_provisioner_role.name
  policy = data.aws_iam_policy_document.ecr_deployment_read_policy.json
}

resource "aws_iam_role_policy" "ftl_admin_ecr_write_access" {
  role   = aws_iam_role.ftl_admin_role.name
  policy = data.aws_iam_policy_document.ecr_deployment_write_policy.json
}
resource "aws_iam_role_policy" "ftl_admin_ecr_read_access" {
  role   = aws_iam_role.ftl_admin_role.name
  policy = data.aws_iam_policy_document.ecr_deployment_read_policy.json
}

data "aws_iam_policy_document" "assume_provisioner_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.kube_namespace}:${local.provisioner_service_account_name}"]
    }
  }
}

resource "aws_iam_role" "ftl_provisioner_role" {
  name               = "${var.ftl_cluster_name}-provisioner-role"
  assume_role_policy = data.aws_iam_policy_document.assume_provisioner_role_policy.json
}

resource "aws_iam_policy" "ftl_provisioner_policy" {
  name        = "${var.ftl_cluster_name}-provisioner-policy"
  description = "Policy to allow FTL provisioner to create infrastructure"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["sts:AssumeRole", "sts:AssumeRoleWithWebIdentity"]
        Effect   = "Allow"
        Resource = aws_iam_role.ftl_provisioner_role.arn
      },
      {
        Action   = [
          "cloudformation:CreateStack",
          "cloudformation:DeleteStack",
          "cloudformation:DescribeStacks",
          "cloudformation:UpdateStack",
          "cloudformation:CreateChangeSet",
          "cloudformation:DescribeChangeSet",
          "cloudformation:DeleteChangeSet",
          "cloudformation:ExecuteChangeSet",
        ]
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action   = [
          "rds:DescribeDBClusters",
          "rds:CreateDBCluster",
          "rds:AddTagsToResource",
          "rds:DescribeDBInstances",
          "rds:DescribeDBSubnetGroups",
          "rds:DescribeDBSecurityGroups",
          "rds:DescribeDBClusterParameterGroups",
          "rds:CreateDBInstance",
          "rds:CreateDBClusterSnapshot",
          "rds:ModifyDBCluster",
          "rds:DeleteDBCluster",

          "secretsmanager:CreateSecret",
          "secretsmanager:TagResource",

          "kms:DescribeKey",
        ]
        Effect   = "Allow"
        Resource = "*"
      }, {
        # Allow the provisioner to get the secret value for the RDS clusters belonging to the FTL cluster.
        # This is needed for creating databases and running DB migrations.
        Action   = [
          "secretsmanager:GetSecretValue",
        ]
        Effect   = "Allow"
        Resource = "*"
        "Condition": {
          "StringEquals": {"aws:ResourceTag/aws:secretsmanager:owningService": "rds"},
          "StringEquals": {"aws:ResourceTag/ftl:cluster": var.ftl_cluster_name}
        }
      }, {
        # Allows to connect to DBs for running migrations with IAM authentication
        "Action": [
            "rds-db:connect"
        ]
        Effect   = "Allow"
        Resource = "*"
      }, {
        # Allows management of Kafka clusters and topics
        "Effect": "Allow",
        "Action": [
            "kafka-cluster:Connect",
            "kafka-cluster:AlterCluster",
            "kafka-cluster:DescribeCluster",
            "kafka-cluster:*Topic*",
            "kafka-cluster:WriteData",
            "kafka-cluster:ReadData",
            "kafka-cluster:AlterGroup",
            "kafka-cluster:DescribeGroup"
        ],
        "Resource": "*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ftl_provisioner_role_policy_attachment" {
  role       = aws_iam_role.ftl_provisioner_role.name
  policy_arn = aws_iam_policy.ftl_provisioner_policy.arn
}

# TODO: These should be created by the FTL provisioner.
data "aws_iam_policy_document" "assume_runner_role_policy" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}"]
    }

    condition {
      test     = "StringLike"
      variable = "${replace(data.aws_eks_cluster.eks.identity[0].oidc[0].issuer, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.kube_namespace}:*"]
    }
  }
}

resource "aws_iam_policy" "ftl_runner_policy" {
  name        = "${var.ftl_cluster_name}-runner-policy"
  description = "Policy to allow FTL runner to connect to infrastructure"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["sts:AssumeRole", "sts:AssumeRoleWithWebIdentity"]
        Effect   = "Allow"
        Resource = aws_iam_role.ftl_runner_role.arn
      }, {
        "Action": [
             "rds-db:connect"
        ]
        Effect   = "Allow"
        Resource = "*"
      }, {
        "Effect": "Allow",
        "Action": [
            "kafka-cluster:Connect",
            "kafka-cluster:WriteData",
            "kafka-cluster:ReadData",
        ],
        "Resource": "*"
      }
    ]
  })
}

resource "aws_iam_role" "ftl_runner_role" {
  name               = "${var.ftl_cluster_name}-runner-role"
  assume_role_policy = data.aws_iam_policy_document.assume_runner_role_policy.json
}

resource "aws_iam_role_policy_attachment" "ftl_runner_role_policy_attachment" {
  role       = aws_iam_role.ftl_runner_role.name
  policy_arn = aws_iam_policy.ftl_runner_policy.arn
}
