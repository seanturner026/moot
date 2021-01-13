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
	"github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type listReposEvent struct {
	RepoOwner string `json:"repo_owner"`
}

type reposList struct {
	RepoOwner  string `dynamodbav:"repo_owner" json:"repo_owner"`
	RepoName   string `dynamodbav:"pk" json:"repo_name"`
	BranchBase string `dynamodbav:"branch_base" json:"branch_base"`
	BranchHead string `dynamodbav:"branch_head" json:"branch_head"`
}

type application struct {
	config configuration
}

type configuration struct {
	GlobalSecondaryIndexName string
	TableName                string
	db                       dynamodbiface.DynamoDBAPI
}

func (app *application) listRepos(e listReposEvent) (dynamodb.QueryOutput, error) {
	input := &dynamodb.QueryInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":type": {
				S: aws.String("repo"),
			},
			":repo_owner": {
				S: aws.String(e.RepoOwner),
			},
		},
		IndexName:              aws.String(app.config.GlobalSecondaryIndexName),
		KeyConditionExpression: aws.String("sk = :type AND repo_owner = :repo_owner"),
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

func (app *application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := listReposEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	output, err := app.listRepos(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to query repos belonging to %v, %v", e.RepoOwner, err), 404, err, headers)
		return resp, nil
	}

	reposList := []reposList{}
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &reposList)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to unmarshal DynamoDB resp, %v", err), 404, err, headers)
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

	log.Printf("[DEBUG] body %v", buf.String())
	resp := events.APIGatewayProxyResponse{
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
