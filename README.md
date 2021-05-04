### Serverless Release Dashboard

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

#### Repositories View

![alt text](https://github.com/seanturner026/serverless-release-dashboard/blob/main/assets/repositories.png?raw=true)

#### Add Repository View
![alt text](https://github.com/seanturner026/serverless-release-dashboard/blob/main/assets/repositories-add.png?raw=true)

#### Users View

![alt text](https://github.com/seanturner026/serverless-release-dashboard/blob/main/assets/users.png?raw=true)
