locals {
  path = "${path.root}/../.."

  ssm_parameters = {
    client_pool_secret = {
      description     = "Cognito User Pool client secret."
      parameter_value = aws_cognito_user_pool_client.this.client_secret
    }
    github_token = {
      description     = "Token for releasing on Github.com."
      parameter_value = var.github_token
    }
    slack_webhook_url = {
      description     = "URL to send slack message payloads to."
      parameter_value = var.slack_webhook_url
    }
  }

  lambdas = {
    auth = {
      description = "Administrates user login, token refreshes, and password resets."
      authorizer  = false
      environment = {
        CLIENT_POOL_ID     = aws_cognito_user_pool_client.this.id
        CLIENT_POOL_SECRET = aws_ssm_parameter.this["client_pool_secret"].value
        USER_POOL_ID       = aws_cognito_user_pool.this.id
      }
      routes = {
        "/auth/login"          = "POST"
        "/auth/refresh/token"  = "POST"
        "/auth/reset/password" = "POST"
      }
      iam_statements = {
        cognito = {
          actions = [
            "cognito-idp:AdminRespondToAuthChallenge",
            "cognito-idp:InitiateAuth",
          ]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }

    releases_create = {
      description = "Creates github releases for repository specified in the event."
      authorizer  = true
      environment = {
        GITHUB_TOKEN      = aws_ssm_parameter.this["github_token"].value
        SLACK_WEBHOOK_URL = aws_ssm_parameter.this["slack_webhook_url"].value
        TABLE_NAME        = aws_dynamodb_table.this.id
      }
      routes = {
        "/releases/create" = "POST"
      }
      iam_statements = {
        dynamodb = {
          actions   = ["dynamodb:UpdateItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
      }
    }

    repositories = {
      description = "Writes github repository details to DynamoDB."
      authorizer  = true
      environment = {
        GLOBAL_SECONDARY_INDEX_NAME = var.global_secondary_index_name
        TABLE_NAME                  = aws_dynamodb_table.this.id
      }
      routes = {
        "/repositories/create" = "POST"
        "/repositories/delete" = "POST"
        "/repositories/list"   = "POST"
      }
      iam_statements = {
        dynamodb = {
          actions = [
            "dynamodb:BatchWriteItem",
            "dynamodb:PutItem",
            "dynamodb:Query",
            "dynamodb:UpdateItem",
          ]
          resources = [
            aws_dynamodb_table.this.arn,
            "${aws_dynamodb_table.this.arn}/index/${var.global_secondary_index_name}",
          ]
        }
      }
    }

    users = {
      description = "Creates, Lists, and Deletes Cognito Users."
      authorizer  = true
      environment = {
        REGION       = data.aws_region.this.name
        USER_POOL_ID = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/create" = "POST"
        "/users/delete" = "POST"
        "/users/list"   = "GET"
      }
      iam_statements = {
        cognito = {
          actions = [
            "cognito-idp:AdminCreateUser",
            "cognito-idp:AdminDeleteUser",
            "cognito-idp:ListUsers"
          ]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }
  }

  lambdas_flat = flatten([
    for lambda_key, lambda_value in local.lambdas : [
      for route, method in lambda_value.routes : {
        lambda_key = lambda_key
        authorizer = lambda_value.authorizer
        method     = method
        route      = route
      }
    ]
  ])
  lambda_integrations = {
    for integration in local.lambdas_flat : "${integration.lambda_key}.${integration.route}.${integration.method}" => integration
  }
}
