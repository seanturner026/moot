package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

type listUsersResponse struct {
	Users []user
}

type user struct {
	Email string `json:"name"`
	ID    string `json:"id"`
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
	for _, u := range users {
		userName := user{
			Email: *u.Attributes[0].Value,
			ID:    *u.Username,
		}
		userNames.appendUserToResponse(userName)
	}
	return *userNames
}

func (userNames *listUsersResponse) appendUserToResponse(u user) {
	userNames.Users = append(userNames.Users, u)
}

func (app application) usersListHandler(event events.APIGatewayV2HTTPRequest) (string, int) {

	listUsersResp, err := app.listUsers()
	if err != nil || len(listUsersResp.Users) == 0 {
		message := "Unable to query list of users"
		statusCode := 400
		return message, statusCode
	}

	userNames := generateListUsersResponse(listUsersResp.Users)

	body, err := json.Marshal(userNames.Users)
	statusCode := 200
	if err != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", err)
		statusCode = 400
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	return buf.String(), statusCode
}
