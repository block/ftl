# Specify the required version of OpenTofu
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.0"
    }
  }
}

# Configure the AWS provider
provider "aws" {
  region = var.region
}

data "aws_eks_cluster" "eks_cluster" {
  name = var.eks_cluster_name
}

data "aws_eks_cluster_auth" "eks_cluster_auth" {
  name = var.eks_cluster_name
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.eks_cluster.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.eks_cluster.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.eks_cluster_auth.token
  }
  burst_limit = 1000
}

# First create the AWS resources
module "ftl_aws" {
  source = "../../modules/ftl-aws"
  
  eks_cluster_name  = var.eks_cluster_name
  ftl_cluster_name = var.ftl_cluster_name
  kube_namespace   = var.kube_namespace
}

# Then create the Helm deployment using the AWS resources
module "ftl" {
  source = "../../modules/ftl"
  
  eks_cluster_name  = var.eks_cluster_name
  ftl_cluster_name = var.ftl_cluster_name
  kube_namespace   = var.kube_namespace
  ftl_version      = var.ftl_version
  chart_values     = [file("${path.module}/${var.helm_values_file}"), module.sandbox.values]
  
  # AWS resource references from ftl-aws module
  controller_role_arn              = module.ftl_aws.controller_role_arn
  provisioner_role_arn            = module.ftl_aws.provisioner_role_arn
  admin_role_arn                  = module.ftl_aws.admin_role_arn
  runner_role_arn                 = module.ftl_aws.runner_role_arn
  kms_key_arn                     = module.ftl_aws.kms_key_arn
  ecr_repository_url              = module.ftl_aws.ecr_repository_url
  controller_service_account_name = module.ftl_aws.controller_service_account_name
  provisioner_service_account_name = module.ftl_aws.provisioner_service_account_name
  admin_service_account_name      = module.ftl_aws.admin_service_account_name
}

module "sandbox" {
  source = "../../modules/sandbox"
  ftl_cluster_name = var.ftl_cluster_name
  vpc_name = var.vpc_name
}