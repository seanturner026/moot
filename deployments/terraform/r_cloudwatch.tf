resource "aws_cloudwatch_log_group" "this" {
  for_each = merge(local.lambdas, { "api_gateway" = {} })

  name              = "/aws/lambda/${var.tags.name}_${each.key}"
  retention_in_days = 7
  tags              = var.tags
}
