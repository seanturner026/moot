resource "aws_route53_record" "alias" {
  count = var.hosted_zone_name != "" && var.fqdn_alias != "" ? 1 : 0

  zone_id = data.aws_route53_zone.this[0].zone_id
  name    = var.fqdn_alias
  type    = "A"

  alias {
    name                   = module.cloudfront.cloudfront_distribution_domain_name
    zone_id                = module.cloudfront.cloudfront_distribution_hosted_zone_id
    evaluate_target_health = false
  }
}

resource "aws_route53_record" "acm" {
  for_each = var.hosted_zone_name != "" && var.fqdn_alias != "" ? {
    for dvo in aws_acm_certificate.this[0].domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  } : {}

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.this[0].zone_id
}
