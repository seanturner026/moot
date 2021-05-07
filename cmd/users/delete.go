package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type deleteUserEvent struct {
	EmailAddress string `json:"email_address"`
}

func (app application) deleteUserFromCognito(e deleteUserEvent) error {
	input := &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
		Username:   aws.String(e.EmailAddress),
	}
	_, err := app.Config.IDP.AdminDeleteUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return err
	}
	log.Info(fmt.Sprintf("deleted user %v successfully", e.EmailAddress))
	return nil
}

func (app application) deleteUserFromDynamoDB(e deleteUserEvent) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("user"),
			},
			"SK": {
				S: aws.String(e.EmailAddress),
			},
		},
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.Config.TableName),
	}
	_, err := app.Config.DB.DeleteItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return err
	}
	log.Info(fmt.Sprintf("deleted user %v successfully", e.EmailAddress))
	return nil
}

func (app application) usersDeleteHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := deleteUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	err = app.deleteUserFromCognito(e)
	if err != nil {
		message := fmt.Sprintf("Error deleting user %v, please refresh and try again.", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	err = app.deleteUserFromDynamoDB(e)
	if err != nil {
		message := fmt.Sprintf("Error deleting user %v from backend, please refresh and try again.", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Deleted user %v", e.EmailAddress)
	statusCode := 200
	return message, statusCode
}
