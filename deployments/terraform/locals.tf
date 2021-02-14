locals {
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
    releases_create = {
      description = "test"
      environment = {
        GITHUB_TOKEN      = var.github_token
        SLACK_WEBHOOK_URL = var.slack_webhook_url
      }
      routes = {
        "/release/create" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        dynamodb = {
          actions   = ["dynamodb:UpdateItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
      }
    }

    repositories_create = {
      description = "test"
      environment = { TABLE_NAME = aws_dynamodb_table.this.id }
      routes = {
        "/repositories/create" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        dynamodb = {
          actions   = ["dynamodb:PutItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
      }
    }

    repositories_delete = {
      description = "test"
      environment = { TABLE_NAME = aws_dynamodb_table.this.id }
      routes = {
        "/repositories/delete" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        dynamodb = {
          actions   = ["dynamodb:PutItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
      }
    }

    repositories_list = {
      description = "test"
      environment = {
        GLOBAL_SECONDARY_INDEX_NAME = var.global_secondary_index_name
        TABLE_NAME                  = aws_dynamodb_table.this.id
      }
      routes = {
        "/repositories/list" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        dynamodb = {
          actions = ["dynamodb:Query"]
          resources = [
            aws_dynamodb_table.this.arn,
            "${aws_dynamodb_table.this.arn}/index/repos",
          ]
        }
      }
    }

    users_create = {
      description = "test"
      environment = {
        REGION       = data.aws_region.this.name
        USER_POOL_ID = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/create" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        cognito = {
          actions   = ["cognito-idp:AdminCreateUser"]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }

    users_delete = {
      description = "test"
      environment = {
        USER_POOL_ID = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/delete" = {
          method     = "POST"
          authorizer = true
        }
      }
      iam_statements = {
        cognito = {
          actions   = ["cognito-idp:AdminDeleteUser"]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }

    users_list = {
      description = "test"
      environment = {
        USER_POOL_ID = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/list" = {
          method     = "GET"
          authorizer = true
        }
      }
      iam_statements = {
        cognito = {
          actions   = ["cognito-idp:ListUsers"]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }

    users_login = {
      description = "test"
      environment = {
        CLIENT_POOL_ID     = aws_cognito_user_pool_client.this.id
        CLIENT_POOL_SECRET = aws_ssm_parameter.this["client_pool_secret"].value
        USER_POOL_ID       = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/login" = {
          method     = "POST"
          authorizer = false
        }
        "/users/refresh/token" = {
          method     = "POST"
          authorizer = false
        }
      }
      iam_statements = {
        cognito = {
          actions   = ["cognito-idp:InitiateAuth"]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }

    users_reset_password = {
      description = "test"
      environment = {
        CLIENT_POOL_ID     = aws_cognito_user_pool_client.this.id
        CLIENT_POOL_SECRET = aws_ssm_parameter.this["client_pool_secret"].value
        USER_POOL_ID       = aws_cognito_user_pool.this.id
      }
      routes = {
        "/users/reset/password" = {
          method     = "POST"
          authorizer = false
        }
      }
      iam_statements = {
        cognito = {
          actions   = ["cognito-idp:AdminRespondToAuthChallenge"]
          resources = [aws_cognito_user_pool.this.arn]
        }
      }
    }
  }

  lambdas_flat = flatten([
    for lambda_key, lambda_value in local.lambdas : [
      for route_key, route_value in lambda_value.routes : {
        lambda_key = lambda_key
        authorizer = route_value.authorizer
        method     = route_value.method
        route      = route_key
      }
    ]
  ])
  lambda_integrations = { for integration in local.lambdas_flat : "${integration.lambda_key}.${integration.route}.${integration.method}" => integration }
}
