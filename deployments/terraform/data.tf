data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_route53_zone" "this" {
  count = var.hosted_zone_name != "" ? 1 : 0

  name         = var.hosted_zone_name
  private_zone = false
}

data "external" "admin_user_id" {
  count      = var.admin_user_email != "" && !var.enable_delete_admin_user ? 1 : 0
  depends_on = [null_resource.create_admin_user[0]]

  program = [
    "go", "run", "${path.module}/assets/cognito.go",
    "--admin-user-email", var.admin_user_email,
    "--user-pool-id", aws_cognito_user_pool.this.id,
  ]
}

data "archive_file" "this" {
  for_each   = local.lambdas
  depends_on = [null_resource.lambda_build]

  type        = "zip"
  source_file = "${local.path}/bin/${each.key}"
  output_path = "${local.path}/archive/${each.key}.zip"
}

data "aws_iam_policy_document" "role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "policy" {
  for_each = local.lambdas

  dynamic "statement" {
    for_each = each.value.iam_statements
    iterator = s

    content {
      actions   = s.value.actions
      resources = s.value.resources
    }
  }
}

data "aws_iam_policy_document" "s3" {
  statement {
    actions   = ["s3:GetObject"]
    resources = ["arn:aws:s3:::${replace(var.name, "_", "-")}-${data.aws_caller_identity.current.account_id}/*"]

    principals {
      type        = "AWS"
      identifiers = module.cloudfront.cloudfront_origin_access_identity_iam_arns
    }
  }
}
