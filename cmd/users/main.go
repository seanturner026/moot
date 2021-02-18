package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type application struct {
	config configuration
}

type configuration struct {
	UserPoolID string
	idp        cidpif.CognitoIdentityProviderAPI
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/users/create" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.usersCreateHandler(event, headers)
		return resp

	} else if event.RawPath == "/users/delete" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.usersDeleteHandler(event, headers)
		return resp

	} else if event.RawPath == "/users/list" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.usersListHandler(event, headers)
		return resp

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
		return resp
	}
}

func main() {
	config := configuration{
		UserPoolID: os.Getenv("USER_POOL_ID"),
		idp:        cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
