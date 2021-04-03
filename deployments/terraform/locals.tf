locals {
  path = "${path.root}/../.."

  ssm_parameters = {
    client_pool_secret = {
      description     = "Cognito User Pool client secret."
      parameter_value = aws_cognito_user_pool_client.this.client_secret
    }
    github_token = {
      path            = "/deploy-tower/12345/github/token"
      description     = "Token for Github.com access."
      parameter_value = var.github_token
    }
    gitlab_token = {
      path            = "/deploy-tower/12345/gitlab/token"
      description     = "Token for Gitlab.com access."
      parameter_value = var.gitlab_token
    }
    slack_webhook_url = {
      description     = "URL to send slack message payloads to."
      parameter_value = var.slack_webhook_url
    }
  }

  lambdas = {
    onboarding = {
      description = "Administrates the application."
      authorizer  = true
      environment = {
        ACCOUNT_ID                   = data.aws_caller_identity.current.account_id
        REGION                       = data.aws_region.current.name
        LAMBDA_AUTH_ROLE_ARN         = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/dev_release_dashboard_auth"
        LAMBDA_RELEASES_ROLE_ARN     = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/dev_release_dashboard_releases"
        LAMBDA_REPOSITORIES_ROLE_ARN = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/dev_release_dashboard_repositories"
        LAMBDA_USERS_ROLE_ARN        = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/dev_release_dashboard_users"
        SLACK_WEBHOOK_URL            = aws_ssm_parameter.this["slack_webhook_url"].value
        TABLE_ARN                    = aws_dynamodb_table.this.arn
        TABLE_NAME                   = aws_dynamodb_table.this.id
        USER_POOL_ARN                = aws_cognito_user_pool.this.arn
        USER_POOL_ID                 = aws_cognito_user_pool.this.id
      }
      routes = {
        "/onboarding/create" = "POST"
      }
      iam_statements = {
        all = {
          actions   = ["ssm:AddTagsToResource"]
          resources = ["*"]
        }
        cognito = {
          actions   = ["cognito-idp:AdminAddUserToGroup", "cognito-idp:CreateGroup", "cognito-idp:AdminCreateUser"]
          resources = [aws_cognito_user_pool.this.arn]
        }
        dynamodb = {
          actions   = ["dynamodb:UpdateItem", "dynamodb:PutItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
        iam = {
          actions   = ["iam:CreateRole", "iam:PassRole", "iam:PutRolePolicy", "iam:TagRole"]
          resources = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/LambdaExecutionRole-*"]
        }
        ssm = {
          actions   = ["ssm:PutParameter"]
          resources = ["arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/deploy-tower/*/*/token"]
        }
      }
    }
    auth = {
      description = "Administrates user login, token refreshes, and password resets."
      authorizer  = false
      environment = {
        CLIENT_POOL_ID     = aws_cognito_user_pool_client.this.id
        CLIENT_POOL_SECRET = aws_ssm_parameter.this["client_pool_secret"].value
        TABLE_NAME         = aws_dynamodb_table.this.id
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
        dynamodb = {
          actions   = ["dynamodb:GetItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
      }
    }

    releases = {
      description = "Creates github releases for repository specified in the event."
      authorizer  = true
      environment = {
        SLACK_WEBHOOK_URL = aws_ssm_parameter.this["slack_webhook_url"].value
        TABLE_NAME        = aws_dynamodb_table.this.id
      }
      routes = {
        "/releases/create/github" = "POST"
        "/releases/create/gitlab" = "POST"
      }
      iam_statements = {
        dynamodb = {
          actions   = ["dynamodb:UpdateItem"]
          resources = [aws_dynamodb_table.this.arn]
        }
        ssm = {
          actions   = ["ssm:GetParameter"]
          resources = [aws_ssm_parameter.this["github_token"].arn, aws_ssm_parameter.this["gitlab_token"].arn]
        }
      }
    }

    repositories = {
      description = "Writes github repository details to DynamoDB."
      authorizer  = true
      environment = {
        TABLE_NAME = aws_dynamodb_table.this.id
      }
      routes = {
        "/repositories/create" = "POST"
        "/repositories/delete" = "POST"
        "/repositories/list"   = "GET"
      }
      iam_statements = {
        dynamodb = {
          actions = [
            "dynamodb:BatchWriteItem",
            "dynamodb:PutItem",
            "dynamodb:Query",
            "dynamodb:UpdateItem",
          ]
          resources = [aws_dynamodb_table.this.arn]
        }
        ssm = {
          actions   = ["ssm:GetParameter"]
          resources = [aws_ssm_parameter.this["github_token"].arn, aws_ssm_parameter.this["gitlab_token"].arn]
        }
      }
    }

    users = {
      description = "Creates, Lists, and Deletes Cognito Users."
      authorizer  = true
      environment = {
        REGION       = data.aws_region.current.name
        TABLE_NAME   = aws_dynamodb_table.this.id
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
          ]
          resources = [aws_cognito_user_pool.this.arn]
        }
        dynamodb = {
          actions = [
            "dynamodb:DeleteItem",
            "dynamodb:PutItem",
            "dynamodb:Query",
          ]
          resources = [aws_dynamodb_table.this.arn]
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
