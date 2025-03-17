
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