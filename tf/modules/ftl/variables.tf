variable "eks_cluster_name" {
  type        = string
  description = "The name of the FTL cluster"
}

variable "chart_values" {
  type = list(string)
  description = "The Helm values for the environment"
}

variable "ftl_version" {
    description = "The version of FTL to install"
    type        = string
}

variable "kube_namespace" {
    description = "The namespace to install FTL into"
    type        = string
    default     = "ftl"
}

variable "ftl_cluster_name" {
  description = "The FTL cluster name"
  type = string
}

# AWS resource references
variable "controller_role_arn" {
  description = "ARN of the FTL controller IAM role"
  type = string
}

variable "provisioner_role_arn" {
  description = "ARN of the FTL provisioner IAM role"
  type = string
}

variable "admin_role_arn" {
  description = "ARN of the FTL admin IAM role"
  type = string
}

variable "runner_role_arn" {
  description = "ARN of the FTL runner IAM role"
  type = string
}

variable "kms_key_arn" {
  description = "ARN of the FTL KMS key"
  type = string
}

variable "ecr_repository_url" {
  description = "URL of the FTL ECR repository"
  type = string
}

variable "controller_service_account_name" {
  description = "Name of the controller service account"
  type = string
}

variable "provisioner_service_account_name" {
  description = "Name of the provisioner service account"
  type = string
}

variable "admin_service_account_name" {
  description = "Name of the admin service account"
  type = string
}