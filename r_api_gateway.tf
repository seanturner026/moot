resource "aws_apigatewayv2_api" "this" {
  name          = var.name
  protocol_type = "HTTP"
  description   = "HTTP API for moot, a serverless release dashboard"

  cors_configuration {
    allow_credentials = true
    allow_headers     = ["Content-Type", "Authorization", "X-Session-Id"]
    allow_methods     = ["GET", "OPTIONS", "POST"]
    allow_origins     = ["https://${var.fqdn_alias}"]
    max_age           = 600
  }

  tags = var.tags
}

resource "aws_apigatewayv2_stage" "this" {
  name        = "$default"
  api_id      = aws_apigatewayv2_api.this.id
  auto_deploy = true

  dynamic "access_log_settings" {
    for_each = var.enable_api_gateway_access_logs ? [var.enable_api_gateway_access_logs] : []

    content {
      destination_arn = aws_cloudwatch_log_group.api_gateway[0].arn
      format = jsonencode({
        "requestId" : "$context.requestId",
        "ip" : "$context.identity.sourceIp",
        "requestTime" : "$context.requestTime",
        "httpMethod" : "$context.httpMethod",
        "routeKey" : "$context.routeKey",
        "status" : "$context.status",
        "protocol" : "$context.protocol",
        "responseLength" : "$context.responseLength",
        "integrationError " : "$context.integrationErrorMessage"
      })
    }
  }

  tags = var.tags
}

resource "aws_apigatewayv2_domain_name" "this" {
  count      = var.hosted_zone_name != "" && var.fqdn_alias != "" ? 1 : 0
  depends_on = [aws_acm_certificate_validation.this[0]]

  domain_name = var.fqdn_alias

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.this[0].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = var.tags
}

resource "aws_apigatewayv2_authorizer" "this" {
  name             = var.name
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

