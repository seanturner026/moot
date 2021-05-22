module "vuejs_frontend" {
  source = "github.com/seanturner026/moot-frontend.git?ref=v0.3.0"
}

module "cloudfront" {
  source     = "terraform-aws-modules/cloudfront/aws"
  version    = "v2.4.0"
  depends_on = [aws_acm_certificate_validation.this[0]]

  aliases                       = var.fqdn_alias != "" ? [var.fqdn_alias] : null
  comment                       = "Moot, a Serverless Release Dashboard"
  enabled                       = true
  is_ipv6_enabled               = true
  price_class                   = "PriceClass_All"
  retain_on_delete              = false
  wait_for_deployment           = true
  default_root_object           = "/index.html"
  create_origin_access_identity = true

  origin_access_identities = {
    s3 = "Cloudfront access to moot S3 bucket"
  }

  origin = {
    s3 = {
      domain_name = aws_s3_bucket.this.bucket_domain_name
      s3_origin_config = {
        origin_access_identity = "s3"
      }
    }
  }

  default_cache_behavior = {
    target_origin_id       = "s3"
    viewer_protocol_policy = "redirect-to-https"

    allowed_methods = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods  = ["HEAD", "GET", "OPTIONS"]
    compress        = true
    query_string    = true
  }

  viewer_certificate = {
    cloudfront_default_certificate = var.hosted_zone_name != "" && var.fqdn_alias != "" ? false : true
    minimum_protocol_versione      = var.hosted_zone_name != "" && var.fqdn_alias != "" ? "TLSv1.2" : "TLSv1"
    acm_certificate_arn            = var.hosted_zone_name != "" && var.fqdn_alias != "" ? aws_acm_certificate.this[0].arn : null
    ssl_support_method             = var.hosted_zone_name != "" && var.fqdn_alias != "" ? "sni-only" : null
  }

  custom_error_response = {
    403 = {
      error_caching_min_ttl = 10
      error_code            = 403
      response_code         = 200
      response_page_path    = "/index.html"
    }
    404 = {
      error_caching_min_ttl = 10
      error_code            = 404
      response_code         = 200
      response_page_path    = "/index.html"
    }
  }

  tags = var.tags
}
