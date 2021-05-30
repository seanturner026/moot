output "cloudfront_fqdn" {
  description = <<-DESC
  Domain assigned to cloudfront to access moot. This isn't needed if you are using route53 to
  create an ALIAS record for moot.
  DESC
  value       = module.moot.cloudfront_domain_name
}
