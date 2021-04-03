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
	_, err := app.config.IDP.AdminDeleteUser(input)
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

func (app application) deleteUserFromDynamoDB(e deleteUserEvent, tenantID string) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("org#%s#user", tenantID)),
			},
			"SK": {
				S: aws.String(e.EmailAddress),
			},
		},
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.config.TableName),
	}
	_, err := app.config.DB.DeleteItem(input)
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

func (app application) usersDeleteHandler(event events.APIGatewayV2HTTPRequest, tenantID string) (string, int) {
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

	err = app.deleteUserFromDynamoDB(e, tenantID)
	if err != nil {
		message := fmt.Sprintf("Error deleting user %v from backend, please refresh and try again.", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Deleted user %v", e.EmailAddress)
	statusCode := 200
	return message, statusCode
}
