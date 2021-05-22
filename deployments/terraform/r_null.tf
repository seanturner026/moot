resource "null_resource" "build_frontend" {

  triggers = {
    deploy_on_version_changes = split("?ref=", local.frontend_module_comprehension.Source)[1]
  }

  provisioner "local-exec" {
    command = "cd ${local.frontend_module_path} && yarn install"
  }

  provisioner "local-exec" {
    command = "cd ${local.frontend_module_path} && echo \"VUE_APP_API_GATEWAY_ENDPOINT=${aws_apigatewayv2_api.this.api_endpoint}\" > .env"
  }

  provisioner "local-exec" {
    command = "cd ${local.frontend_module_path} && yarn build"
  }

  provisioner "local-exec" {
    command = "cd ${local.frontend_module_path} && aws s3 sync --cache-control 'max-age=604800' dist/ s3://${aws_s3_bucket.this.id}"
  }
}
