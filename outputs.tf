output "cloudfront_domain_name" {
  description = "FQDN of Cloudfront Distribution that can be used for DNS."
  value       = module.cloudfront.cloudfront_distribution_domain_name
}
