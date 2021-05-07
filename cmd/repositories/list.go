package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	log "github.com/sirupsen/logrus"
)

func (app awsController) listRepos() (dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":primary_key": {
				S: aws.String("repo"),
			}},
		KeyConditionExpression: aws.String("PK = :primary_key"),
		Select:                 aws.String("ALL_ATTRIBUTES"),
		TableName:              aws.String(app.TableName),
	}

	resp, err := app.DB.Query(input)
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

func (r *repository) removeDynamoRepoPartion() {
	repoDetails := strings.Split(r.RepoProvider, "#")
	r.RepoProvider = repoDetails[0]
	r.RepoName = repoDetails[1]
}

func (app application) repositoriesListHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	output, err := app.AWS.listRepos()
	if err != nil {
		message := "Failed to query repositories"
		statusCode := 400
		return message, statusCode
	}

	repos := []*repository{}
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &repos)
	if err != nil {
		message := "Failed to read repositories response"
		statusCode := 400
		return message, statusCode
	}

	for _, repo := range repos {
		repo.removeDynamoRepoPartion()
	}

	body, err := json.Marshal(repos)
	statusCode := 200
	if err != nil {
		log.Error(fmt.Sprintf("unable to marshal json for response, %v", err))
		statusCode = 400
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	return buf.String(), statusCode
}
