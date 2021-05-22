resource "aws_s3_bucket" "this" {
  bucket        = "${replace(var.name, "_", "-")}-${data.aws_caller_identity.current.account_id}"
  acl           = "public-read"
  policy        = data.aws_iam_policy_document.s3.json
  force_destroy = true

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "POST"]
    allowed_origins = [var.fqdn_alias]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }

  website {
    index_document = "index.html"
    error_document = "index.html"
  }

  tags = var.tags
}
