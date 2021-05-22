resource "aws_acm_certificate" "this" {
  count = var.hosted_zone_name != "" && var.fqdn_alias != "" ? 1 : 0

  domain_name       = var.fqdn_alias
  validation_method = "DNS"
  tags              = merge(var.tags, { Name = var.name })

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "this" {
  count = var.hosted_zone_name != "" && var.fqdn_alias != "" ? 1 : 0

  certificate_arn         = aws_acm_certificate.this[0].arn
  validation_record_fqdns = [for record in aws_route53_record.acm : record.fqdn]
}
