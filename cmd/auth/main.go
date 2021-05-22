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
	"github.com/seanturner026/moot/internal/util"
	log "github.com/sirupsen/logrus"
)

type application struct {
	Config configuration
	DB     dynamodbiface.DynamoDBAPI
	IDP    cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

type configuration struct {
	TableName        string
	ClientPoolID     string
	UserPoolID       string
	ClientPoolSecret string
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/auth/login" || event.RawPath == "/auth/refresh/token" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode, headers := app.authLoginHandler(event, headers)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/auth/reset/password" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode, headers := app.authResetPasswordHandler(event, headers)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else {
		log.Error(fmt.Sprintf("path %v does not exist", event.RawPath))
		return util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{}), nil
	}
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := configuration{
		TableName:        os.Getenv("TABLE_NAME"),
		ClientPoolID:     os.Getenv("CLIENT_POOL_ID"),
		UserPoolID:       os.Getenv("USER_POOL_ID"),
		ClientPoolSecret: os.Getenv("CLIENT_POOL_SECRET"),
	}

	app := application{
		Config: config,
		DB:     dynamodb.New(session.Must(session.NewSession())),
		IDP:    cognitoidentityprovider.New(session.Must(session.NewSession())),
	}

	lambda.Start(app.handler)
}
