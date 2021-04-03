package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

func (app application) createUser(e createUserEvent, tenantID string) (string, error) {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:             aws.String(app.config.UserPoolID),
		Username:               aws.String(e.EmailAddress),
		DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.EmailAddress),
			},
			{
				Name:  aws.String("custom:tenant_id"),
				Value: aws.String(tenantID),
			},
		},
	}
	resp, err := app.config.idp.AdminCreateUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}
	log.Printf("[INFO] Created new user %v successfully", e.EmailAddress)
	userID := *resp.User.Username
	return userID, nil
}

func (app application) writeUserToDynamoDB(e createUserEvent, tenantID, userID string) error {
	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("org#%s#user", tenantID)),
			},
			"SK": {
				S: aws.String(e.EmailAddress),
			},
			"Enabled": {
				BOOL: aws.Bool(true),
			},
			"ID": {
				S: aws.String(userID),
			},
		},
		ReturnConsumedCapacity:      aws.String("NONE"),
		ReturnItemCollectionMetrics: aws.String("NONE"),
		ReturnValues:                aws.String("NONE"),
		TableName:                   aws.String(app.config.TableName),
	}
	_, err := app.config.db.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	return nil
}

func (app application) usersCreateHandler(event events.APIGatewayV2HTTPRequest, tenantID string) (string, int) {
	e := createUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	userID, err := app.createUser(e, tenantID)
	if err != nil {
		message := fmt.Sprintf("Error creating user account for %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	// note(SMT): This needs to send a notification if it fails
	err = app.writeUserToDynamoDB(e, tenantID, userID)
	if err != nil {
		message := fmt.Sprintf("Error creating user account for %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Created user account for %v", e.EmailAddress)
	statusCode := 200
	return message, statusCode
}
