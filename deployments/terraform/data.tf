data "aws_region" "this" {}

data "null_data_source" "wait_for_lambda_build" {
  for_each = local.lambdas

  inputs = {
    lambda_build_id = null_resource.lambda_build[each.key].id
    source          = "${local.path}/bin/${each.key}"
  }
}

data "archive_file" "this" {
  for_each = local.lambdas

  type        = "zip"
  source_file = data.null_data_source.wait_for_lambda_build[each.key].outputs["source"]
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
