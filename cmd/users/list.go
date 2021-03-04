package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type listUsersResponse struct {
	Users []userName
}

type userName struct {
	Name string `json:"name"`
}

func (app application) listUsers() (cidp.ListUsersOutput, error) {
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
		return *resp, err
	}

	// NOTE(SMT): Need to implement pagination
	return *resp, err
}

func generateListUsersResponse(users []*cidp.UserType) listUsersResponse {
	userNames := &listUsersResponse{}
	for _, user := range users {
		userName := userName{Name: *user.Attributes[0].Value}
		userNames.appendUserToResponse(userName)
	}
	return *userNames
}

func (userNames *listUsersResponse) appendUserToResponse(user userName) {
	userNames.Users = append(userNames.Users, user)
}

func (app application) usersListHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) events.APIGatewayV2HTTPResponse {

	listUsersResp, err := app.listUsers()
	if err != nil || len(listUsersResp.Users) == 0 {
		resp := util.GenerateResponseBody("Unable to populate list of users", 404, err, headers, []string{})
		return resp
	}

	userNames := generateListUsersResponse(listUsersResp.Users)
	body, err := json.Marshal(userNames.Users)
	statusCode := 200
	if err != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", err)
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	resp := events.APIGatewayV2HTTPResponse{
		StatusCode:      statusCode,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: false,
	}
	return resp
}
