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
	ClientPoolID     string
	UserPoolID       string
	ClientPoolSecret string
	idp              cidpif.CognitoIdentityProviderAPI
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/auth/login" || event.RawPath == "/auth/refresh/token" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.authLoginHandler(event, headers)
		return resp, nil

	} else if event.RawPath == "/auth/reset/password" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.authResetPasswordHandler(event, headers)
		return resp, nil

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
		return resp, nil
	}
}

func main() {
	config := configuration{
		ClientPoolID:     os.Getenv("CLIENT_POOL_ID"),
		UserPoolID:       os.Getenv("USER_POOL_ID"),
		ClientPoolSecret: os.Getenv("CLIENT_POOL_SECRET"),
		idp:              cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
