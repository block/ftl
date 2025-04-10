variable "eks_cluster_name" {
  type        = string
  description = "The name of the FTL cluster"
}

variable "ftl_cluster_name" {
  description = "The FTL cluster name"
  type = string
}

variable "kube_namespace" {
    description = "The namespace to install FTL into"
    type        = string
    default     = "ftl"
}