resource "null_resource" "build_frontend" {

  triggers = {
    deploy_on_version_changes = split("?ref=", local.frontend_module_comprehension.Source)[1]
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.frontend_module_path} && yarn install"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.frontend_module_path} && echo \"VUE_APP_API_GATEWAY_ENDPOINT=${aws_apigatewayv2_api.this.api_endpoint}\" > .env"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.frontend_module_path} && yarn build"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.frontend_module_path} && AWS_DEFAULT_PROFILE=${local.aws_profile} aws s3 sync --cache-control 'max-age=604800' dist/ s3://${aws_s3_bucket.this.id}"
  }
}

resource "null_resource" "go_setup" {

  triggers = {
    hash_go_mod = filemd5("${local.main_module_path}/go.mod")
    hash_go_sum = filemd5("${local.main_module_path}/go.sum")
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cp -f ${local.main_module_path}/go.mod ."
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cp -f ${local.main_module_path}/go.sum ."
  }
}

resource "null_resource" "lambda_build" {
  for_each   = local.lambdas
  depends_on = [null_resource.go_setup]

  triggers = {
    binary_exists = local.null.lambda_binary_exists[each.key]

    hash_main = join("", [
      for file in fileset("${path.module}/cmd/${each.key}", "*.go") : filemd5("${path.module}/cmd/${each.key}/${file}")
    ])

    hash_util = join("", [
      for file in fileset("${path.module}/internal/util", "*.go") : filemd5("${path.module}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "export GO111MODULE=on"
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.main_module_path} && GOOS=linux go build -ldflags '-s -w' -o ./bin/${each.key} ./cmd/${each.key}/."
  }
}

resource "null_resource" "lambda_test" {
  for_each = local.lambdas

  triggers = {
    hash_main = join("", [
      for file in fileset("${path.module}/cmd/${each.key}", "*.go") : filemd5("${path.module}/cmd/${each.key}/${file}")
    ])

    hash_util = join("", [
      for file in fileset("${path.module}/internal/util", "*.go") : filemd5("${path.module}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    interpreter = ["/bin/bash", "-c"]
    command     = "cd ${local.main_module_path} && go test ./cmd/${each.key}"
  }
}

resource "null_resource" "create_admin_user" {
  count = var.admin_user_email != "" && !var.enable_delete_admin_user ? 1 : 0

  provisioner "local-exec" {
    command = "aws --region ${data.aws_region.current.name} cognito-idp admin-create-user --user-pool-id ${aws_cognito_user_pool.this.id} --username ${var.admin_user_email} --user-attributes Name=email,Value=${var.admin_user_email}"
  }
}

resource "null_resource" "delete_admin_user" {
  count = var.admin_user_email != "" && var.enable_delete_admin_user ? 1 : 0

  provisioner "local-exec" {
    command = "aws --region ${data.aws_region.current.name} cognito-idp admin-delete-user --user-pool-id ${aws_cognito_user_pool.this.id} --username ${var.admin_user_email}"
  }
}
