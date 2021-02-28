variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources"
}

variable "global_secondary_index_name" {
  type        = string
  description = "Name of DynamoDB global secondary index."
}

variable "github_token" {
  type        = string
  description = "Token for releasing on Github.com."
}

variable "gitlab_token" {
  type        = string
  description = "Token for releasing on Gitlab.com."
}

variable "slack_webhook_url" {
  type        = string
  description = "URL to send slack message payloads to."
  default     = ""
}
