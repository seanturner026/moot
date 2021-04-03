package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type application struct {
	config configuration
}

type configuration struct {
	TableName  string
	UserPoolID string
	db         dynamodbiface.DynamoDBAPI
	idp        cognitoidentityprovideriface.CognitoIdentityProviderAPI
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}
	IDToken := event.Headers["x-identity-token"]
	tenantID := util.ExtractTenantID(IDToken)

	if event.RawPath == "/users/create" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.usersCreateHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/users/delete" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.usersDeleteHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/users/list" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.usersListHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
		return resp, nil
	}
}

func main() {
	config := configuration{
		TableName:  os.Getenv("TABLE_NAME"),
		UserPoolID: os.Getenv("USER_POOL_ID"),
		db:         dynamodb.New(session.Must(session.NewSession())),
		idp:        cognitoidentityprovider.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
