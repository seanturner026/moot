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
| github\_token | Token for releasing on Github.com. | `string` | n/a | yes |
| global\_secondary\_index\_name | name of DynamoDB global secondary index. | `string` | n/a | yes |
| slack\_webhook\_url | URL to send slack message payloads to. | `string` | n/a | yes |
| tags | Map of tags to be applied to resources | `map(string)` | n/a | yes |

## Outputs

No output.
