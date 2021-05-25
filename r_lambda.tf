resource "aws_lambda_function" "this" {
  depends_on = [null_resource.lambda_build, null_resource.lambda_test]
  for_each   = local.lambdas

  filename         = "${path.module}/archive/${each.key}.zip"
  function_name    = "${var.name}_${each.key}"
  description      = each.value.description
  role             = aws_iam_role.this[each.key].arn
  handler          = each.key
  publish          = false
  source_code_hash = data.archive_file.this[each.key].output_base64sha256
  runtime          = "go1.x"
  timeout          = "10"
  tags             = var.tags

  environment {
    variables = each.value.environment
  }
}

resource "aws_lambda_permission" "this" {
  for_each = local.lambda_integrations

  statement_id  = "AllowAPIGatewayV2Invoke-${replace(replace(each.key, "/", ""), ".", "")}"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[each.value.lambda_key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*${each.value.route}"
}
