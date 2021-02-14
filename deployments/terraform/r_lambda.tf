resource "null_resource" "lambda_build" {
  for_each = local.lambdas

  triggers = {
    main = concat([
      for file in fileset("${path.root}/../../cmd/${each.key}", "*.go") : filebase64("${path.root}/../../cmd/${each.key}/${file}")
    ])[0]
    util = concat([
      for file in fileset("${path.root}/../../internal/util", "*.go") : filebase64("${path.root}/../../internal/util/${file}")
    ])[0]
  }

  provisioner "local-exec" {
    command = "export GO111MODULE=on"
  }

  provisioner "local-exec" {
    command = "GOOS=linux go build -ldflags '-s -w' -o ${path.root}/../../bin/${each.key} ${path.root}/../../cmd/${each.key}/main.go"
  }
}

resource "null_resource" "lambda_test" {
  for_each = local.lambdas

  triggers = {
    main = concat([
      for file in fileset("${path.root}/../../cmd/${each.key}", "*.go") : filebase64("${path.root}/../../cmd/${each.key}/${file}")
    ])[0]
    util = concat([
      for file in fileset("${path.root}/../../internal/util", "*.go") : filebase64("${path.root}/../../internal/util/${file}")
    ])[0]
  }

  provisioner "local-exec" {
    command = "go test ${path.root}/../../cmd/${each.key}"
  }
}

resource "aws_lambda_function" "this" {
  depends_on = [null_resource.lambda_build, null_resource.lambda_test]
  for_each   = local.lambdas

  filename         = "${path.root}/../../archive/${each.key}.zip"
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
  for_each = local.lambdas

  statement_id  = "AllowAPIGatewayV2Invoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.this[each.key].function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.this.execution_arn}/*/*/*"
}
