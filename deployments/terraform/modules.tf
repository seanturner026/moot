module "cloudfront" {
  source  = "terraform-aws-modules/cloudfront/aws"
  version = "v1.8.0"

  # aliases             = [""]
  comment             = "Serverless Release Dashboard"
  enabled             = true
  is_ipv6_enabled     = true
  price_class         = "PriceClass_All"
  retain_on_delete    = false
  wait_for_deployment = false
  default_root_object = "/index.html"

  create_origin_access_identity = true
  origin_access_identities = {
    s3 = "Cloudfront access to Serverless Release Dashboard bucket"
  }

  # logging_config = {
  #   bucket = "logs-my-cdn.s3.amazonaws.com"
  # }

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
    viewer_protocol_policy = "allow-all"

    allowed_methods = ["HEAD", "DELETE", "POST", "GET", "OPTIONS", "PUT", "PATCH"]
    cached_methods  = ["HEAD", "GET", "OPTIONS"]
    compress        = true
    query_string    = true
  }

  viewer_certificate = {
    cloudfront_default_certificate = true,
    minimum_protocol_versione      = "TLSv1"
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
