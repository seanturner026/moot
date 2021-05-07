package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"
)

type user struct {
	Email string `dynamodbav:"SK" json:"name"`
	ID    string `dynamodbav:"ID" json:"id"`
}

func (app application) listUsers() (dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":primary_key": {
				S: aws.String("user"),
			}},
		KeyConditionExpression: aws.String("PK = :primary_key"),
		Select:                 aws.String("ALL_ATTRIBUTES"),
		TableName:              aws.String(app.Config.TableName),
	}

	resp, err := app.Config.DB.Query(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return *resp, err
	}
	return *resp, err
}

func (app application) usersListHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	output, err := app.listUsers()
	if err != nil {
		message := "Unable to query list of users"
		statusCode := 400
		return message, statusCode
	}

	users := []*user{}
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &users)
	if err != nil {
		message := "Failed to read repositories response"
		statusCode := 400
		return message, statusCode
	}

	body, err := json.Marshal(users)
	statusCode := 200
	if err != nil {
		log.Error(fmt.Sprintf("unable to marshal json for response, %v", err))
		statusCode = 400
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	return buf.String(), statusCode
}
