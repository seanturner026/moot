resource "null_resource" "lambda_build" {
  for_each = local.lambdas

  triggers = {
    binary_exists = fileexists("${local.path}/bin/${each.key}")

    main = join("", [
      for file in fileset("${local.path}/cmd/${each.key}", "*.go") : filebase64("${local.path}/cmd/${each.key}/${file}")
    ])

    util = join("", [
      for file in fileset("${local.path}/internal/util", "*.go") : filebase64("${local.path}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    command = "export GO111MODULE=on"
  }

  provisioner "local-exec" {
    command = "GOOS=linux go build -ldflags '-s -w' -o ${local.path}/bin/${each.key} ${local.path}/cmd/${each.key}/."
  }
}

resource "null_resource" "lambda_test" {
  for_each = local.lambdas

  triggers = {
    main = join("", [
      for file in fileset("${local.path}/cmd/${each.key}", "*.go") : filebase64("${local.path}/cmd/${each.key}/${file}")
    ])

    util = join("", [
      for file in fileset("${local.path}/internal/util", "*.go") : filebase64("${local.path}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    command = "go test ${local.path}/cmd/${each.key}"
  }
}

resource "aws_lambda_function" "this" {
  depends_on = [null_resource.lambda_build, null_resource.lambda_test]
  for_each   = local.lambdas

  filename         = "${local.path}/archive/${each.key}.zip"
  function_name    = each.key
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
