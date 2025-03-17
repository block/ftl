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
    value = aws_iam_role.ftl_controller_role.arn
  }

  set {
    name  = "controller.kmsUri"
    value = "aws-kms://${aws_kms_key.ftl_db_kms_key.arn}"
  }

  set {
    name  = "provisioner.provisionersRoleArn"
    value = aws_iam_role.ftl_provisioner_role.arn
  }

  set {
    name  = "provisioner.serviceAccountName"
    value = local.provisioner_service_account_name
  }

  set {
    name  = "admin.adminRoleArn"
    value = aws_iam_role.ftl_admin_role.arn
  }

  set {
    name  = "runner.runnersRoleArn"
    value = aws_iam_role.ftl_runner_role.arn
  }

  set {
    name  = "controller.serviceAccount"
    value = local.controller_service_account_name
  }

  set {
    name  = "admin.serviceAccountName" #TODO: better service account handling
    value = local.admin_service_account_name
  }
  set {
    name  = "registry.repository"
    value = aws_ecr_repository.ftl_deployment_content_repo.repository_url
  }
}
