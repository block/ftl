data aws_eks_cluster eks {
  name = var.eks_cluster_name
}

resource "helm_release" "ftl" {
  name             = var.ftl_cluster_name
  chart            = "ftl"
  repository       = "https://block.github.io/ftl-charts/"
  namespace        = var.kube_namespace
  create_namespace = true
  timeout          = 180
  version = var.ftl_version
  upgrade_install = true

  values = var.chart_values

  set {
    name  = "controller.controllersRoleArn"
    value = var.controller_role_arn
  }

  set {
    name  = "controller.kmsUri"
    value = "aws-kms://${var.kms_key_arn}"
  }

  set {
    name  = "provisioner.provisionersRoleArn"
    value = var.provisioner_role_arn
  }

  set {
    name  = "provisioner.serviceAccountName"
    value = var.provisioner_service_account_name
  }

  set {
    name  = "admin.adminRoleArn"
    value = var.admin_role_arn
  }

  set {
    name  = "runner.runnersRoleArn"
    value = var.runner_role_arn
  }

  set {
    name  = "controller.serviceAccount"
    value = var.controller_service_account_name
  }

  set {
    name  = "admin.serviceAccountName"
    value = var.admin_service_account_name
  }
  
  set {
    name  = "registry.repository"
    value = var.ecr_repository_url
  }
}