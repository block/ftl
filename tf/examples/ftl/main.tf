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
data "aws_eks_cluster_auth" "cluster" {
  name = var.eks_cluster_name
}

module ftl {
  source            = "../../modules/ftl"
  eks_cluster_name  = var.eks_cluster_name
  chart_values = [file("${path.module}/${var.helm_values_file}"), module.sandbox.values]
  ftl_version       = var.ftl_version
  ftl_cluster_name = var.ftl_cluster_name
  kube_namespace = "test"
}

module sandbox {
  source = "../../modules/sandbox"
  ftl_cluster_name = var.ftl_cluster_name
  vpc_name = var.vpc_name
}
