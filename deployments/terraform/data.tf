data "aws_region" "this" {}

data "null_data_source" "wait_for_lambda_build" {
  for_each = local.lambdas

  inputs = {
    lambda_build_id = null_resource.lambda_build[each.key].id
    lambda_test_id  = null_resource.lambda_test[each.key].id
    source          = "${path.root}/../../bin/${each.key}"
  }
}

data "archive_file" "this" {
  for_each = local.lambdas

  type        = "zip"
  source_file = data.null_data_source.wait_for_lambda_build[each.key].outputs["source"]
  output_path = "${path.root}/../../archive/${each.key}.zip"
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
