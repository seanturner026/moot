output "cloudfront_fqdn" {
  description = "Domain assigned to cloudfront to access moot."
  value       = module.moot.cloudfront_domain_name
}
