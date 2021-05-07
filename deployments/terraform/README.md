## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| archive | n/a |
| aws | n/a |
| null | n/a |
| random | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| admin\_user\_email | Email address for dashboard admin | `string` | n/a | yes |
| dev\_cloudfront\_dns | n/a | `string` | n/a | yes |
| enable\_admin\_user\_creation | Controls the creation of an admin user that is required to initially gain access to the<br>dashboard.<br><br>If access to the dashboard is completely lost, do the following<br>• `var.enable_admin_user_creation = false`<br>• terraform apply<br>• `var.enable_admin_user_creation = true`<br>• terraform apply<br><br>If the initial admin user should no longer be able to access the dashboard, revoke access by<br>setting `var.enable_admin_user_creation = false` and running `terraform apply` | `bool` | `true` | no |
| github\_token | Token for Github. | `string` | `""` | no |
| gitlab\_token | Token for Gitlab. | `string` | `""` | no |
| slack\_webhook\_url | URL to send slack message payloads to. | `string` | `""` | no |
| tags | Map of tags to be applied to resources. | `map(string)` | n/a | yes |

## Outputs

No output.
