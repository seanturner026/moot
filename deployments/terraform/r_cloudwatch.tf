resource "aws_cloudwatch_log_group" "this" {
  for_each = merge(local.lambdas, { "api_gateway" = {} })

  name              = "/aws/lambda/${each.key}"
  retention_in_days = 7
  tags              = var.tags
}
