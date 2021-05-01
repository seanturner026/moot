variable "tags" {
  type        = map(string)
  description = "Map of tags to be applied to resources."
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

variable "stripe_token" {
  type        = string
  description = "Token for Stripe."
}

variable "dev_cloudfront_dns" {
  type = string
}
