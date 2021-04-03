variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources"
}

variable "github_token" {
  type        = string
  description = "Token for Github.com."
  default     = ""
}

variable "gitlab_token" {
  type        = string
  description = "Token for Gitlab.com."
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
