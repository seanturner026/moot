module "moot" {
  source = "github.com/seanturner026/moot.git"

  name                           = "moot"
  aws_profile                    = "default"
  admin_user_email               = var.admin_user_email
  enable_delete_admin_user       = false
  github_token                   = var.github_token
  gitlab_token                   = var.gitlab_token
  slack_webhook_url              = var.slack_webhook_url
  enable_api_gateway_access_logs = true
  tags                           = var.tags
}
