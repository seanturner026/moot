resource "aws_cognito_user_pool" "this" {
  name                     = var.tags.name
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  admin_create_user_config {

    invite_message_template {
      email_subject = "Serverless Release Dashboard User Signup"
      email_message = file("${path.root}/assets/cognito_invite_template.html")
      sms_message   = <<-MESSAGE
      username: {username}
      password: {####}
      MESSAGE
    }
  }

  account_recovery_setting {
    recovery_mechanism {
      name     = "admin_only"
      priority = 1
    }
  }

  password_policy {
    minimum_length                   = 36
    require_lowercase                = true
    require_numbers                  = true
    require_symbols                  = false
    require_uppercase                = true
    temporary_password_validity_days = 1
  }

  tags = var.tags
}

resource "aws_cognito_user_pool_client" "this" {
  name                                 = var.tags.name
  user_pool_id                         = aws_cognito_user_pool.this.id
  generate_secret                      = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_scopes                 = ["email", "openid"]
  explicit_auth_flows = [
    "ALLOW_ADMIN_USER_PASSWORD_AUTH",
    "ALLOW_CUSTOM_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_USER_SRP_AUTH",
  ]
  supported_identity_providers = ["COGNITO"]
  callback_urls                = ["https://localhost:3000"]
}

resource "aws_cognito_identity_pool" "this" {
  identity_pool_name               = var.tags.name
  allow_unauthenticated_identities = false

  cognito_identity_providers {
    client_id     = aws_cognito_user_pool_client.this.id
    provider_name = aws_cognito_user_pool.this.endpoint
  }

  tags = var.tags
}

resource "null_resource" "create_admin_user" {
  count      = var.enable_admin_user_creation ? 1 : 0
  depends_on = [aws_lambda_function.this]

  provisioner "local-exec" {
    command = "aws --region ${data.aws_region.current.name} cognito-idp admin-create-user --user-pool-id ${aws_cognito_user_pool.this.id} --username ${var.admin_user_email} --user-attributes Name=email,Value=${var.admin_user_email}"
  }
}
