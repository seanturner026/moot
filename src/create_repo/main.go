package main

import (
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

type createRepoEvent struct {
	RepoOwner  string `dynamodbav:"repo_owner" json:"repo_owner"`
	RepoName   string `dynamodbav:"pk" json:"repo_name"`
	Type       string `dynamodbav:"sk"`
	BranchBase string `dynamodbav:"branch_base" json:"branch_base"`
	BranchHead string `dynamodbav:"branch_head" json:"branch_head"`
}

type application struct {
	config configuration
}

type configuration struct {
	TableName string
	db        dynamodbiface.DynamoDBAPI
}

func generatePutItemInput(e createRepoEvent) (createRepoEvent, map[string]*dynamodb.AttributeValue, error) {
	e.Type = "repo"
	itemInput, err := dynamodbattribute.MarshalMap(e)
	if err != nil {
		return e, map[string]*dynamodb.AttributeValue{}, err
	}
	return e, itemInput, nil
}

func (app *application) writeRepoToDB(e createRepoEvent, itemInput map[string]*dynamodb.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(app.config.TableName),
		Item:                   itemInput,
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
	log.Printf("[INFO] Wrote ID %s successfully", e.RepoName)
	return nil
}

func (app *application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := createRepoEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	e, itemInput, err := generatePutItemInput(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to stage provided information for loading into DynamoDB for ID %v, %v", e.RepoName, err), 404, err, headers)
		return resp, nil
	}
	log.Printf("[DEBUG] event %v", e)
	log.Printf("[DEBUG] input %v", itemInput)
	err = app.writeRepoToDB(e, itemInput)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to write record %v to DynamoDB table, %v", e.RepoName, err), 404, err, headers)
		return resp, nil
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Wrote record %v to DynamoDB successfully", e.RepoName), 200, nil, headers)
	return resp, nil
}

func main() {
	config := configuration{
		TableName: os.Getenv("TABLE_NAME"),
		db:        dynamodb.New(session.Must(session.NewSession())),
	}

	app := application{config: config}
	lambda.Start(app.handler)
}
