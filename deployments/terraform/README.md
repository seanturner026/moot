## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| archive | n/a |
| aws | n/a |
| null | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| account\_id | Target account ID. | `string` | n/a | yes |
| github\_token | Token for releasing on Github.com. | `string` | `""` | no |
| gitlab\_token | Token for releasing on Gitlab.com. | `string` | `""` | no |
| slack\_webhook\_url | URL to send slack message payloads to. | `string` | `""` | no |
| tags | Map of tags to be applied to resources | `map(string)` | n/a | yes |

## Outputs

No output.
