resource "aws_apigatewayv2_api" "this" {
  name          = var.tags.name
  protocol_type = "HTTP"
  description   = "HTTP API for serverless release dashboard"

  cors_configuration {
    allow_credentials = true
    allow_headers     = ["Content-Type", "Authorization", "X-Session-Id"]
    allow_methods     = ["GET", "OPTIONS", "POST", ]
    allow_origins     = ["http://localhost:8080"]
    max_age           = 600
  }

  tags = var.tags
}

resource "aws_apigatewayv2_stage" "this" {
  name        = "$default"
  api_id      = aws_apigatewayv2_api.this.id
  auto_deploy = true
  tags        = var.tags
}

resource "aws_apigatewayv2_authorizer" "this" {
  name             = var.tags.name
  api_id           = aws_apigatewayv2_api.this.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]

  jwt_configuration {
    audience = [aws_cognito_user_pool_client.this.id]
    issuer   = "https://${aws_cognito_user_pool.this.endpoint}"
  }
}

resource "aws_apigatewayv2_integration" "this" {
  for_each = local.lambda_integrations

  api_id                 = aws_apigatewayv2_api.this.id
  integration_type       = "AWS_PROXY"
  connection_type        = "INTERNET"
  integration_method     = each.value.method
  integration_uri        = aws_lambda_function.this[each.value.lambda_key].arn
  timeout_milliseconds   = 10500
  payload_format_version = "2.0"

  # tls_config {
  #   server_name_to_verify = ""
  # }
}

resource "aws_apigatewayv2_route" "this" {
  for_each = local.lambda_integrations

  api_id             = aws_apigatewayv2_api.this.id
  route_key          = "${each.value.method} ${each.value.route}"
  authorization_type = each.value.authorizer ? "JWT" : "NONE"
  authorizer_id      = each.value.authorizer ? aws_apigatewayv2_authorizer.this.id : null
  target             = "integrations/${aws_apigatewayv2_integration.this[each.key].id}"
}

resource "aws_apigatewayv2_route" "proxy" {
  for_each = local.lambda_integrations

  api_id             = aws_apigatewayv2_api.this.id
  route_key          = "OPTIONS ${each.value.route}/{proxy+}"
  authorization_type = each.value.authorizer ? "JWT" : "NONE"
  authorizer_id      = each.value.authorizer ? aws_apigatewayv2_authorizer.this.id : null
  target             = "integrations/${aws_apigatewayv2_integration.this[each.key].id}"
}

