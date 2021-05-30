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
}

type configuration struct {
	TableName  string
	UserPoolID string
	DB         dynamodbiface.DynamoDBAPI
	IDP        cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/users/create" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode := app.usersCreateHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/users/delete" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode := app.usersDeleteHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/users/list" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode := app.usersListHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
	}

	log.Error(fmt.Sprintf("path %v does not exist", event.RawPath))
	resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
	return resp, nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := configuration{
		TableName:  os.Getenv("TABLE_NAME"),
		UserPoolID: os.Getenv("USER_POOL_ID"),
		DB:         dynamodb.New(session.Must(session.NewSession())),
		IDP:        cognitoidentityprovider.New(session.Must(session.NewSession())),
	}

	app := application{Config: config}

	lambda.Start(app.handler)
}
