data "aws_caller_identity" "current" {}
resource "aws_kms_key" "ftl_db_kms_key" {
  description             = "KMS key for encrypting secrets"
  deletion_window_in_days = 7

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "key-policy",
  "Statement": [
    {
      "Sid": "AllowRootAccess",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "kms:*",
      "Resource": "*"
    },
    {
      "Sid": "AllowRoleUseOfKMS",
      "Effect": "Allow",
      "Principal": {
        "AWS": "${aws_iam_role.ftl_admin_role.arn}"
      },
      "Action": [
        "kms:Encrypt",
        "kms:Decrypt",
        "kms:GenerateDataKey*",
        "kms:DescribeKey"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kms_alias" "ftl_db_kms_key_alias" {
  name          = "alias/${var.ftl_cluster_name}-db-kms-key-alias"
  target_key_id = aws_kms_key.ftl_db_kms_key.key_id
}
