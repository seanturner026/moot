variable "name" {
  type        = string
  description = "Name to be applied to all resources."
  default     = "release_dashboard"
}

variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources."
  default     = {}
}

variable "aws_profile" {
  type        = string
  description = <<-DESC
  AWS Profile Name from `~/.aws/config that can be used for local execution. This profile is used
  to preform the following actions:

  • `aws s3 sync`: Sync bundle produced by `yarn` to build to s3
  • `cognito-idp admin-create-user`: Creates an admin cognito user for dashboard access
  • `cognito-idp admin-delete-user`: Deletes an admin cognito user if the user should not 
  have access to the dashboard anymore, OR, if there is no way for the user to regain access.
  • `cognito-idp list-users`: Obtains the admin user's ID in order to write the ID to the 
  DynamodDB table.
  DESC
  default     = ""
}

variable "admin_user_email" {
  type        = string
  description = <<-DESC
  Controls the creation of an admin user that is required to initially gain access to the
  dashboard.

  If access to the dashboard is completely lost, do the following
  • `var.enable_delete_admin_user = true`
  • `terraform apply`
  • `var.enable_delete_admin_user = false`
  • `terraform apply`

  If the initial admin user should no longer be able to access the dashboard, revoke access by
  setting `var.enable_delete_admin_user = true` and running `terraform apply`
  DESC
  default     = ""
}

variable "enable_delete_admin_user" {
  type        = bool
  description = <<-DESC
  Destroys the admin user.

  Set this value to true to destroy the user, and to false to recreate the user.
  DESC
  default     = false
}

variable "github_token" {
  type        = string
  description = "Token for Github."
  default     = "42"
}

variable "gitlab_token" {
  type        = string
  description = "Token for Gitlab."
  default     = "42"
}

variable "slack_webhook_url" {
  type        = string
  description = "URL to send slack message payloads to."
  default     = "42"
}

variable "hosted_zone_name" {
  type        = string
  description = "Name of AWS Route53 Hosted Zone for DNS."
  default     = ""
}

variable "fqdn_alias" {
  type        = string
  description = <<-DESC
  ALIAS for the Cloudfront distribution, S3, Cognito and API Gateway. Must be in the form of
  `example.com`.
  DESC
  default     = ""
}

variable "enable_api_gateway_access_logs" {
  type        = bool
  description = "Enables API Gateway access logging to cloudwatch for the default stage."
  default     = false
}
