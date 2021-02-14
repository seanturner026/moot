package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type listReposEvent struct {
	RepoOwner string `json:"repo_owner"`
}

type reposList struct {
	RepoOwner    string `dynamodbav:"SK" json:"repo_owner"`
	RepoName     string `dynamodbav:"PK" json:"repo_name"`
	RepoProvider string `dynamodbav:"RepoProvider" json:"repo_provider"`
	BranchBase   string `dynamodbav:"BranchBase" json:"branch_base"`
	BranchHead   string `dynamodbav:"BranchHead" json:"branch_head"`
}

type application struct {
	config configuration
}

type configuration struct {
	GlobalSecondaryIndexName string
	TableName                string
	db                       dynamodbiface.DynamoDBAPI
}

func (app application) listRepos(e listReposEvent) (dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":sort_key": {
				S: aws.String("repo#" + e.RepoOwner),
			},
			// ":repo_provider": {
			// 	// NOTE(SMT): This is poorly implemented
			// 	S: aws.String("github.com"), // e.RepoProvider
			// },
		},
		IndexName:              aws.String(app.config.GlobalSecondaryIndexName),
		KeyConditionExpression: aws.String("SK = :sort_key"), //AND repo_provider = :repo_provider"),
		Select:                 aws.String("ALL_ATTRIBUTES"),
		TableName:              aws.String(app.config.TableName),
	}

	resp, err := app.config.db.Query(input)
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

func (app application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := listReposEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	output, err := app.listRepos(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to query repos belonging to %v, %v", e.RepoOwner, err), 404, err, headers, []string{})
		return resp, nil
	}

	reposList := []reposList{}
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &reposList)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to unmarshal DynamoDB resp, %v", err), 404, err, headers, []string{})
		return resp, nil
	}

	body, err := json.Marshal(reposList)
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
	return resp, nil
}

func main() {
	config := configuration{
		GlobalSecondaryIndexName: os.Getenv("GLOBAL_SECONDARY_INDEX_NAME"),
		TableName:                os.Getenv("TABLE_NAME"),
		db:                       dynamodb.New(session.Must(session.NewSession())),
	}

	app := application{config: config}
	lambda.Start(app.handler)
}
