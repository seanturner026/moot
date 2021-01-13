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
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type deleteRepoEvent struct {
	RepoName string `dynamodbav:"pk" json:"repo_name"`
}

type application struct {
	config configuration
}

type configuration struct {
	TableName string
	db        dynamodbiface.DynamoDBAPI
}

func (app *application) deleteRepo(e deleteRepoEvent) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {
				S: aws.String(e.RepoName),
			},
			"sk": {
				S: aws.String("repo"),
			},
		},
		ReturnConsumedCapacity: aws.String("NONE"),
		ReturnValues:           aws.String("NONE"),
		TableName:              aws.String(app.config.TableName),
	}
	_, err := app.config.db.DeleteItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Deleted ID %s successfully", e.RepoName)

	return nil
}

func (app application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := deleteRepoEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	err = app.deleteRepo(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to delete repo %v, %v", e.RepoName, err), 404, err, headers)
		return resp, nil
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Deleted repo %v successfully", e.RepoName), 200, err, headers)
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
