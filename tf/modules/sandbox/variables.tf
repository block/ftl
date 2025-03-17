

variable "ftl_cluster_name" {
  type        = string
  description = "The name of the FTL cluster"
}

variable "vpc_name" {
  description = "Name of the VPC where the cluster security group will be provisioned"
  type        = string
  default     = null
}

