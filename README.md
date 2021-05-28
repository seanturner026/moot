# Moot - A Serverless Release Dashboard

AWS Serverless solution that implements a release dashboard that provides a single button press for deploying code changes to production.

## How it all works

User's onboard github or gitlab repositories (need to specify a BASE (main) and HEAD (develop) branch) in the frontend, at which point you can then `deploy` code changes to a production environment by hitting `deploy`. Deploying creates pull requests which merge the HEAD branch into BASE, and creates a release. Users can also select `hotfix`, which skips the pull request and creates a release based on the BASE branch.

To make it all work, you'll need to configure your production continuous integration deployment pipeline trigger with a regex check on the release version number so that it's only triggered on semver releases, for example. You'll also need to provide a github or gitlab API token to give the dashboard access to make API calls to the respective VCS provider. These tokens will be stored as SSM parameters within AWS.

Deploys trigger the following workflow:
  - Create Github / Gitlab PR base <- head (e.g. main <- develop)
  - Approve PR
  - Create Release based on base branch
  - Send Slack message to a channel with the release notes.

Hotfix Deploys trigger the following workflow:
  - Create Release based on base branch
  - Send Slack message to a channel with the release notes.

This solution utilises the following services:
  - API Gateway (auth + routing)
  - Cloudwatch (logging)
  - Cognito (auth)
  - DynamoDB (backend storage)
  - Lambda (backend compute)
  - S3 + Cloudfront (VueJS frontend)
  - SSM Parameter Store (secrets management)

## Requirements

The following tools must be installed in order to fully deploy Moot

- [yarn](https://yarnpkg.com/getting-started/install) -- used to build the frontend locally
- [awscli](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html) -- used to aws s3 sync the frontend assets
- [go](https://golang.org/doc/install#download) -- requires at least version 16 because of `go modules`
- gcc (build essentials) -- required for `go build`

## Installation

The below snippet will fully deploy the dashboard (backend + frontend). In this instance, I am deploying to a cheap Route53 domain I purchased for testing purposes (`moot.link`).

If your AWS account does not have a Route53 hosted zone, remove the `hosted_zone_name` and `fqdn_alias` lines to use Cloudfront's default certificate and dns.

See the [terraform_examples](https://github.com/seanturner026/moot/tree/main/terraform_examples) directory for deployable examples.

```hcl
module "moot" {
  source = "github.com/seanturner026/moot.git"

  name                           = "moot"
  aws_profile                    = "default"
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

## Repositories View

![alt text](https://github.com/seanturner026/moot/blob/main/assets/repositories.png?raw=true)

## Add Repository View
![alt text](https://github.com/seanturner026/moot/blob/main/assets/repositories-add.png?raw=true)

## Users View

![alt text](https://github.com/seanturner026/moot/blob/main/assets/users.png?raw=true)

## Terraform Information

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| archive | n/a |
| aws | n/a |
| external | n/a |
| null | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| admin\_user\_email | Controls the creation of an admin user that is required to initially gain access to the<br>dashboard.<br><br>If access to the dashboard is completely lost, do the following<br>• `var.enable_delete_admin_user = true`<br>• `terraform apply`<br>• `var.enable_delete_admin_user = false`<br>• `terraform apply`<br><br>If the initial admin user should no longer be able to access the dashboard, revoke access by<br>setting `var.enable_delete_admin_user = true` and running `terraform apply` | `string` | `""` | no |
| aws\_profile | AWS Profile Name from ~/.aws/config that can be used for local execution. This profile is used<br>to preform the following actions:<br><br>• `aws s3 sync`: Sync bundle produced by `yarn` to build to s3<br><br>• `cognito-idp admin-create-user`: Creates an admin cognito user for dashboard access<br><br>• `cognito-idp admin-delete-user`: Deletes an admin cognito user if the user should not<br>have access to the dashboard anymore, OR, if there is no way for the user to regain access.<br><br>• `cognito-idp list-users`: Obtains the admin user's ID in order to write the ID to the<br>DynamodDB table. | `string` | `""` | no |
| enable\_api\_gateway\_access\_logs | Enables API Gateway access logging to cloudwatch for the default stage. | `bool` | `false` | no |
| enable\_delete\_admin\_user | Destroys the admin user.<br><br>Set this value to true to destroy the user, and to false to recreate the user. | `bool` | `false` | no |
| fqdn\_alias | ALIAS for the Cloudfront distribution, S3, Cognito and API Gateway. Must be in the form of<br>`example.com`. | `string` | `""` | no |
| github\_token | Token for Github. | `string` | `"42"` | no |
| gitlab\_token | Token for Gitlab. | `string` | `"42"` | no |
| hosted\_zone\_name | Name of AWS Route53 Hosted Zone for DNS. | `string` | `""` | no |
| name | Name to be applied to all resources. | `string` | `"release_dashboard"` | no |
| slack\_webhook\_url | URL to send slack message payloads to. | `string` | `"42"` | no |
| tags | Map of tags to be applied to resources. | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| cloudfront\_domain\_name | FQDN of Cloudfront Distribution that can be used for DNS. |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
