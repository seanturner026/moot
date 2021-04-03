package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
)

type application struct {
	Config configuration
}

type configuration struct {
	AccountID                 string
	Region                    string
	SlackWebhookURL           string
	TableARN                  string
	TableName                 string
	UserPoolARN               string
	UserPoolID                string
	LambdaAuthRoleARN         string
	LambdaReleasesRoleARN     string
	LambdaRepositoriesRoleArn string
	LambdaUsersRoleArn        string
	DB                        dynamodbiface.DynamoDBAPI
	IAM                       iamiface.IAMAPI
	IDP                       cognitoidentityprovideriface.CognitoIdentityProviderAPI
	SSM                       ssmiface.SSMAPI
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/admin/overview" {
		log.Info(fmt.Sprintf("handling request on %s", event.RawPath))
		message, statusCode := app.overviewHandler()
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else {
		log.Error(fmt.Sprintf("path %s does not exist", event.RawPath))
		return util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{}), nil
	}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := configuration{
		AccountID:                 os.Getenv("ACCOUNT_ID"),
		Region:                    os.Getenv("REGION"),
		SlackWebhookURL:           os.Getenv("SLACK_WEBHOOK_URL"),
		TableARN:                  os.Getenv("TABLE_ARN"),
		TableName:                 os.Getenv("TABLE_NAME"),
		UserPoolARN:               os.Getenv("USER_POOL_ARN"),
		UserPoolID:                os.Getenv("USER_POOL_ID"),
		LambdaAuthRoleARN:         os.Getenv("LAMBDA_AUTH_ROLE_ARN"),
		LambdaReleasesRoleARN:     os.Getenv("LAMBDA_RELEASES_ROLE_ARN"),
		LambdaRepositoriesRoleArn: os.Getenv("LAMBDA_REPOSITORIES_ROLE_ARN"),
		LambdaUsersRoleArn:        os.Getenv("LAMBDA_USERS_ROLE_ARN"),
		DB:                        dynamodb.New(session.Must(session.NewSession())),
		IAM:                       iam.New(session.Must(session.NewSession())),
		IDP:                       cognitoidentityprovider.New(session.Must(session.NewSession())),
		SSM:                       ssm.New(session.Must(session.NewSession())),
	}

	app := application{
		Config: config,
	}

	lambda.Start(app.handler)
}
