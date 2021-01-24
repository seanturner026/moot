import os
import cfnresponse

import boto3

client = boto3.client("cognito-idp")


def handler(event, context):
    if event["RequestType"] == "Create":
        response = client.describe_user_pool_client(
            UserPoolId=os.getenv("USER_POOL_ID"), ClientId=os.getenv("CLIENT_POOL_ID")
        )
        client_secret = response["UserPoolClient"]["ClientSecret"]
        print(client_secret)

        responseData = {}
        responseData["Data"] = client_secret

        cfnresponse.send(event, context, cfnresponse.SUCCESS, responseData)
