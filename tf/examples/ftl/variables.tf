variable "region" {
  description = "The AWS region"
  default     = "us-west-2"
  type        = string
}

variable "helm_values_file" {
  type        = string
  description = "The name of the Helm values file for the environment"
  default     = "values.yaml"
}

variable "ftl_version" {
  description = "The version of FTL to install"
  type        = string
  default     = "0.466.0"
}

variable "ftl_cluster_name" {
  description = "The FTL cluster name"
  type        = string
  default     = "ftl"
}

variable "eks_cluster_name" {
  description = "The EKS cluster name"
  type        = string
  default     = "eks-vpc"
}

variable "vpc_name" {
  description = "The name of the VPC"
  type        = string
  default     = "ftl-vpc"
}

variable "kube_namespace" {
  description = "The namespace to install FTL into"
  type        = string
  default     = "test"
}