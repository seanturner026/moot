resource "aws_cloudwatch_log_group" "this" {
  for_each = local.lambdas

  name = "/${var.tags.name}/${each.key}"
  tags = var.tags
}
