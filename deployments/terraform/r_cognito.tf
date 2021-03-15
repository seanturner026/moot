resource "aws_cognito_user_pool" "this" {
  name                     = var.tags.name
  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  password_policy {
    minimum_length                   = 36
    require_lowercase                = true
    require_numbers                  = true
    require_symbols                  = false
    require_uppercase                = true
    temporary_password_validity_days = 1
  }

  schema {
    name                     = "tenant_id"
    attribute_data_type      = "String"
    mutable                  = true
    required                 = false
    developer_only_attribute = false

    string_attribute_constraints {
      max_length = "256"
      min_length = "1"
    }
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
  read_attributes              = ["custom:tenant_id"]
  write_attributes             = ["custom:tenant_id"]
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
