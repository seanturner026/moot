package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func (app awsController) listRepos(tenantID string) (dynamodb.QueryOutput, error) {
	log.Printf("TenantID %+v", tenantID)
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":primary_key": {
				S: aws.String(fmt.Sprintf("org#%s#repo", tenantID)),
			}},
		KeyConditionExpression: aws.String("PK = :primary_key"),
		Select:                 aws.String("ALL_ATTRIBUTES"),
		TableName:              aws.String(app.TableName),
	}

	resp, err := app.db.Query(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
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

func (app application) repositoriesListHandler(event events.APIGatewayV2HTTPRequest, tenantID string) (string, int) {
	output, err := app.aws.listRepos(tenantID)
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
		log.Printf("[ERROR] Unable to marshal json for response, %v", err)
		statusCode = 400
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	return buf.String(), statusCode
}
