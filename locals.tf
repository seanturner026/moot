locals {
  ssm_parameters = {
    client_pool_secret = {
      description     = "Cognito User Pool client secret."
      parameter_value = aws_cognito_user_pool_client.this.client_secret
    }
    github_token = {
      description     = "Token for Github access."
      parameter_value = var.github_token == "" ? 42 : var.github_token
    }
    gitlab_token = {
      description     = "Token for Gitlab access."
      parameter_value = var.gitlab_token == "" ? 42 : var.gitlab_token
    }
    slack_webhook_url = {
      description     = "URL to send slack message payloads to."
      parameter_value = var.slack_webhook_url == "" ? 42 : var.slack_webhook_url
    }
  }

  null = {
    lambda_binary_exists = { for key, _ in local.lambdas : key => fileexists("${path.module}/bin/${key}") }
  }

  frontend_module_comprehension = [for module in jsondecode(file("${path.root}/.terraform/modules/modules.json"))["Modules"] : module if length(regexall("vuejs_frontend", module.Key)) > 0][0]
  frontend_module_path          = "${path.root}/${local.frontend_module_comprehension.Dir}"
  main_module_path              = "./.terraform/modules/${local.main_module_name}"
  main_module_name              = split(".terraform/modules/", path.module)[1]

  lambdas = {
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
      }
    }

    releases = {
      description = "Creates github and gitlab releases for repository specified in the event."
      authorizer  = true
      environment = {
        DASHBOARD_NAME    = var.name
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
      description = "Writes github and gitlab repository details to DynamoDB."
      authorizer  = true
      environment = {
        DASHBOARD_NAME = var.name
        TABLE_NAME     = aws_dynamodb_table.this.id
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
