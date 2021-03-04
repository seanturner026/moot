resource "aws_s3_bucket" "this" {
  bucket = "release-dashboard-${var.account_id}"
  acl    = "public-read"
  policy = data.aws_iam_policy_document.s3.json

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "POST"]
    allowed_origins = [var.dev_cloudfront_dns]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }

  website {
    index_document = "index.html"
    error_document = "index.html"
  }

  tags = var.tags
}
