package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type deleteUserEvent struct {
	EmailAddress string `json:"email_address"`
}

func (app application) deleteUser(e deleteUserEvent) error {
	input := &cidp.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
		Username:   aws.String(e.EmailAddress),
	}
	_, err := app.config.idp.AdminDeleteUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Deleted user %v successfully", e.EmailAddress)
	return nil
}

func (app application) usersDeleteHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) events.APIGatewayV2HTTPResponse {
	e := deleteUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	err = app.deleteUser(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error deleting user %v", e.EmailAddress), 404, err, headers, []string{})
		return resp
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Deleted user %v", e.EmailAddress), 200, nil, headers, []string{})
	return resp
}
