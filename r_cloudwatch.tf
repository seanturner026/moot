resource "aws_cloudwatch_log_group" "this" {
  for_each = local.lambdas

  name              = "/aws/lambda/${var.name}_${each.key}"
  retention_in_days = 7
  tags              = var.tags
}

resource "aws_cloudwatch_log_group" "api_gateway" {
  count = var.enable_api_gateway_access_logs ? 1 : 0

  name              = "/aws/${var.name}/api_gateway/access"
  retention_in_days = 7
  tags              = var.tags
}
