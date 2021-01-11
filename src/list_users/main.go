package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type application struct {
	config configuration
}

type configuration struct {
	UserPoolID string
	idp        cidpif.CognitoIdentityProviderAPI
}

type listUsersResponse struct {
	Users []userName
}

type userName struct {
	Name string `json:"name"`
}

func (userNames *listUsersResponse) appendUserToResponse(user userName) {
	userNames.Users = append(userNames.Users, user)
}

func (app *application) listUsers() (listUsersResponse, error) {
	input := &cidp.ListUsersInput{
		AttributesToGet: aws.StringSlice([]string{"email"}),
		Limit:           aws.Int64(60),
		UserPoolId:      aws.String(app.config.UserPoolID),
	}

	resp, err := app.config.idp.ListUsers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return listUsersResponse{}, err
	}

	// NOTE(SMT): reflect []*cidp.UserType to stringslice?
	users := resp.Users
	userNames := &listUsersResponse{}
	for _, user := range users {
		userName := userName{Name: *user.Attributes[0].Value}
		userNames.appendUserToResponse(userName)
	}
	return *userNames, nil
}

func (app *application) handler() (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	userNames, err := app.listUsers()
	if err != nil || len(userNames.Users) == 0 {
		resp := util.GenerateResponseBody("Unable to populate list of users", 404, err, headers)
		return resp, nil
	}

	body, marshalErr := json.Marshal(userNames.Users)
	statusCode := 200
	if marshalErr != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", marshalErr)
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	resp := events.APIGatewayProxyResponse{
		StatusCode:      statusCode,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: false,
	}
	return resp, nil
}

func main() {
	config := configuration{
		UserPoolID: os.Getenv("USER_POOL_ID"),
		idp:        cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
