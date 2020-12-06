### Serverless Release Dashboard

Dashboard which will support the following workflow (executed by API Gateway and Lambda) for 1 to many Github Repos:
  - Create Github PR base <- head (e.g. main <- develop)
  - Approve PR
  - Create Release based on base branch
  - Send Slack message to a channel indicating that a release has been created.

The Dashboard will integrate AWS Cognito to lock down the Lambdas
