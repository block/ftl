output "controller_role_arn" {
  description = "ARN of the FTL controller IAM role"
  value = aws_iam_role.ftl_controller_role.arn
}

output "provisioner_role_arn" {
  description = "ARN of the FTL provisioner IAM role"
  value = aws_iam_role.ftl_provisioner_role.arn
}

output "admin_role_arn" {
  description = "ARN of the FTL admin IAM role"
  value = aws_iam_role.ftl_admin_role.arn
}

output "runner_role_arn" {
  description = "ARN of the FTL runner IAM role"
  value = aws_iam_role.ftl_runner_role.arn
}

output "kms_key_arn" {
  description = "ARN of the FTL KMS key"
  value = aws_kms_key.ftl_db_kms_key.arn
}

output "ecr_repository_url" {
  description = "URL of the FTL ECR repository"
  value = aws_ecr_repository.ftl_deployment_content_repo.repository_url
}

output "controller_service_account_name" {
  description = "Name of the controller service account"
  value = local.controller_service_account_name
}

output "provisioner_service_account_name" {
  description = "Name of the provisioner service account"
  value = local.provisioner_service_account_name
}

output "admin_service_account_name" {
  description = "Name of the admin service account"
  value = local.admin_service_account_name
}