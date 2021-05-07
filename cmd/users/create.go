package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

func (app application) createUser(e createUserEvent) (string, error) {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:             aws.String(app.Config.UserPoolID),
		Username:               aws.String(e.EmailAddress),
		DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.EmailAddress),
			},
		},
	}
	resp, err := app.Config.IDP.AdminCreateUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return "", err
	}
	log.Info(fmt.Sprintf("created new user %v successfully", e.EmailAddress))
	userID := *resp.User.Username
	return userID, nil
}

func (app application) writeUserToDynamoDB(e createUserEvent, userID string) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("user"),
			},
			"SK": {
				S: aws.String(e.EmailAddress),
			},
			"ID": {
				S: aws.String(userID),
			},
		},
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.Config.TableName),
	}
	_, err := app.Config.DB.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return err
	}
	return nil
}

func (app application) usersCreateHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := createUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	userID, err := app.createUser(e)
	if err != nil {
		message := fmt.Sprintf("Error creating user account for %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	// note(SMT): This needs to send a notification if it fails
	err = app.writeUserToDynamoDB(e, userID)
	if err != nil {
		message := fmt.Sprintf("Error creating user account for %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Created user account for %v", e.EmailAddress)
	statusCode := 200
	return message, statusCode
}
