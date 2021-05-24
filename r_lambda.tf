resource "null_resource" "go_setup" {

  triggers = {
    hash_go_mod = filemd5("${local.main_module_path}/go.mod")
    hash_go_sum = filemd5("${local.main_module_path}/go.sum")
  }

  provisioner "local-exec" {
    command = "cp -f ${local.main_module_path}/go.mod ."
  }

  provisioner "local-exec" {
    command = "cp -f ${local.main_module_path}/go.sum ."
  }
}

resource "null_resource" "lambda_build" {
  for_each   = local.lambdas
  depends_on = [null_resource.go_setup]

  triggers = {
    binary_exists = local.null.lambda_binary_exists[each.key]

    hash_main = join("", [
      for file in fileset("${path.module}/cmd/${each.key}", "*.go") : filemd5("${path.module}/cmd/${each.key}/${file}")
    ])

    hash_util = join("", [
      for file in fileset("${path.module}/internal/util", "*.go") : filemd5("${path.module}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    command = "export GO111MODULE=on"
  }

  provisioner "local-exec" {
    command = "cd ${local.main_module_path} && GOOS=linux go build -ldflags '-s -w' -o ./bin/${each.key} ./cmd/${each.key}/."
  }
}

resource "null_resource" "lambda_test" {
  for_each = local.lambdas

  triggers = {
    hash_main = join("", [
      for file in fileset("${path.module}/cmd/${each.key}", "*.go") : filemd5("${path.module}/cmd/${each.key}/${file}")
    ])

    hash_util = join("", [
      for file in fileset("${path.module}/internal/util", "*.go") : filemd5("${path.module}/internal/util/${file}")
    ])
  }

  provisioner "local-exec" {
    command = "cd ${local.main_module_path} && go test ./cmd/${each.key}"
  }
}

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
