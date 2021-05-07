variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources."
}

variable "admin_user_email" {
  type        = string
  description = "Email address for dashboard admin"
}

variable "enable_admin_user_creation" {
  type        = bool
  description = <<-DESC
  Controls the creation of an admin user that is required to initially gain access to the
  dashboard.

  If access to the dashboard is completely lost, do the following
  • `var.enable_admin_user_creation = false`
  • terraform apply
  • `var.enable_admin_user_creation = true`
  • terraform apply

  If the initial admin user should no longer be able to access the dashboard, revoke access by
  setting `var.enable_admin_user_creation = false` and running `terraform apply`
  DESC
  default     = true
}

variable "github_token" {
  type        = string
  description = "Token for Github."
  default     = ""
}

variable "gitlab_token" {
  type        = string
  description = "Token for Gitlab."
  default     = ""
}

variable "slack_webhook_url" {
  type        = string
  description = "URL to send slack message payloads to."
  default     = ""
}

variable "dev_cloudfront_dns" {
  type = string
}
