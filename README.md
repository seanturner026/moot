### Moot - A Serverless Release Dashboard

AWS Serverless solution deployed with Terraform which implements a single-page-application dashboard. This dashboard creates releases that are intended to trigger continuous integration (CI) production deploy pipelines. All that is needed to kick off a release is a version number.

Deploys trigger the following workflow:
  - Create Github / Gitlab PR base <- head (e.g. main <- develop)
  - Approve PR
  - Create Release based on base branch
  - Send Slack message to a channel indicating that a release has been created.

This solution utilises the following services:
  - API Gateway (auth + routing)
  - Cloudwatch (logging)
  - Cognito (auth)
  - DynamoDB (backend storage)
  - Lambda (backend compute)
  - S3 + Cloudfront (frontend)
  - SSM Parameter Store (secrets management)

#### Installation

```hcl
module "moot" {
  source = "github.com/seanturner026/moot.git"

  name                           = "moot"
  admin_user_email               = var.admin_user_email
  enable_delete_admin_user       = false
  github_token                   = var.github_token
  gitlab_token                   = var.gitlab_token
  slack_webhook_url              = var.slack_webhook_url
  fqdn_alias                     = "moot.link"
  hosted_zone_name               = "moot.link"
  enable_api_gateway_access_logs = true
  tags                           = var.tags
}
```

#### Workflows

- Standard Deploy: Merges the HEAD branch into the BASE (e.g. main) branch, creates release based on BASE branch
- Hotfix Deploy: Creates release based on the BASE branch


#### Repositories View

![alt text](https://github.com/seanturner026/moot/blob/main/assets/repositories.png?raw=true)

#### Add Repository View
![alt text](https://github.com/seanturner026/moot/blob/main/assets/repositories-add.png?raw=true)

#### Users View

![alt text](https://github.com/seanturner026/moot/blob/main/assets/users.png?raw=true)

## Terraform Providers

| Name | Version |
|------|---------|
| archive | n/a |
| aws | n/a |
| external | n/a |
| null | n/a |

## Terraform Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| admin\_user\_email | Controls the creation of an admin user that is required to initially gain access to the<br>dashboard.<br><br>If access to the dashboard is completely lost, do the following<br>• `var.enable_delete_admin_user = true`<br>• `terraform apply`<br>• `var.enable_delete_admin_user = false`<br>• `terraform apply`<br><br>If the initial admin user should no longer be able to access the dashboard, revoke access by<br>setting `var.enable_delete_admin_user = true` and running `terraform apply` | `string` | `""` | no |
| enable\_api\_gateway\_access\_logs | Enables API Gateway access logging to cloudwatch for the default stage. | `bool` | `false` | no |
| enable\_delete\_admin\_user | Destroys the admin user.<br><br>Set this value to true to destroy the user, and to false to recreate the user. | `bool` | `false` | no |
| fqdn\_alias | ALIAS for the Cloudfront distribution, S3, Cognito and API Gateway. Must be in the form of<br>`example.com`. | `string` | `""` | no |
| github\_token | Token for Github. | `string` | `""` | no |
| gitlab\_token | Token for Gitlab. | `string` | `""` | no |
| hosted\_zone\_name | Name of AWS Route53 Hosted Zone for DNS. | `string` | `""` | no |
| name | Name to be applied to all resources. | `string` | `"release_dashboard"` | no |
| slack\_webhook\_url | URL to send slack message payloads to. | `string` | `""` | no |
| tags | Map of tags to be applied to resources. | `map(string)` | `{}` | no |

## Terraform Outputs

| Name | Description |
|------|-------------|
| cloudfront\_domain\_name | FQDN of Cloudfront Distribution that can be used for DNS. |
