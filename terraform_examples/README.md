### Complete Example

This is a full invocation and provides everything the module has to offer.

### Cloudfront DNS Example

This example creates everything which is created by the complete example. However, it does not create ACM or Route53 resources, and uses the default cloudfront DNS and certficate. The Terraform module invocation is not extenable such that it is possible to provide an ACM certificate ARN and alias to utilise non-default cloudfront DNS and certificate.
