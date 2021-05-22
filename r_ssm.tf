resource "aws_ssm_parameter" "this" {
  for_each = local.ssm_parameters

  name        = "/${var.name}/${each.key}"
  description = each.value.description
  type        = "SecureString"
  value       = each.value.parameter_value
  tags        = var.tags
}
