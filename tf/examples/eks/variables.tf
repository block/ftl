
variable "region" {
  description = "The AWS region"
  default = "us-west-2"
  type = string
}

variable "vpc_name" {
  default = "ftl-vpc"
  description = "The name of the VPC"
  type = string
}
variable "eks_cluster_name" {
  default = "ftl-eks"
  description = "The name of the EKS cluster"
  type = string
}

variable "vpc_cidr" {
  default = "10.0.0.0/16"
    description = "The CIDR block for the VPC"
    type        = string
}

variable "eks_instance_types" {
  description = "List of instance types for the EKS node groups"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_ami_type" {
  description = "AMI type for the EKS node groups"
  type        = string
  default     = "AL2023_x86_64_STANDARD"
}

variable "eks_node_group_min_size" {
  description = "Minimum number of nodes in the EKS node group"
  type        = number
  default     = 3
}

variable "eks_node_group_max_size" {
  description = "Maximum number of nodes in the EKS node group"
  type        = number
  default     = 9
}

variable "eks_node_group_desired_size" {
  description = "Desired number of nodes in the EKS node group"
  type        = number
  default     = 4
}