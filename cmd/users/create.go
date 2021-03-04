package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

func (app application) createUser(e createUserEvent) error {
	input := &cidp.AdminCreateUserInput{
		UserPoolId:             aws.String(app.config.UserPoolID),
		Username:               aws.String(e.EmailAddress),
		DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cidp.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.EmailAddress),
			},
		},
	}
	_, err := app.config.idp.AdminCreateUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Created new user %v successfully", e.EmailAddress)
	return nil
}

func (app application) usersCreateHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) events.APIGatewayV2HTTPResponse {
	e := createUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	err = app.createUser(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error creating user %v", e.EmailAddress), 404, err, headers, []string{})
		return resp
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Created new user %v", e.EmailAddress), 200, nil, headers, []string{})
	return resp
}
