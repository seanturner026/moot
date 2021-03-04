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
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
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

func (app application) repositoriesListHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) events.APIGatewayV2HTTPResponse {
	output, err := app.aws.listRepos()
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to query repos, %v", err), 404, err, headers, []string{})
		return resp
	}

	repos := []*repository{}
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &repos)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to unmarshal DynamoDB resp, %v", err), 404, err, headers, []string{})
		return resp
	}

	for _, repo := range repos {
		repo.removeDynamoRepoPartion()
	}

	body, err := json.Marshal(repos)
	statusCode := 200
	if err != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", err)
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	resp := events.APIGatewayV2HTTPResponse{
		StatusCode:      statusCode,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: false,
	}
	return resp
}
