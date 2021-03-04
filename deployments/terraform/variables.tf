variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources"
}

variable "account_id" {
  type        = string
  description = "Target account ID."
}

variable "global_secondary_index_name" {
  type        = string
  description = "Name of DynamoDB global secondary index."
}

variable "github_token" {
  type        = string
  description = "Token for releasing on Github.com."
  default     = ""
}

variable "gitlab_token" {
  type        = string
  description = "Token for releasing on Gitlab.com."
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
