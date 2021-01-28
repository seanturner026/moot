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
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type deleteRepositoriesEvent struct {
	Repositories []string `json:"repositories"`
}

type application struct {
	config configuration
}

type configuration struct {
	TableName string
	db        dynamodbiface.DynamoDBAPI
}

func (app application) stageBatchWrites(e deleteRepositoriesEvent) error {
	repositories := []*dynamodb.WriteRequest{}
	for i, r := range e.Repositories {
		if (i+1)%25 == 0 {
			app.deleteRepositories(repositories)
			repositories = []*dynamodb.WriteRequest{}
		}
		deleteRequest := &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: map[string]*dynamodb.AttributeValue{
					"pk": {
						S: aws.String(r),
					},
					"sk": {
						S: aws.String("repo"),
					},
				},
			},
		}
		repositories = append(repositories, deleteRequest)
	}
	if len(repositories) != 0 {
		app.deleteRepositories(repositories)
	}
	log.Println("[INFO] Deleted IDs successfully")
	return nil
}

func (app application) deleteRepositories(requestItems []*dynamodb.WriteRequest) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			os.Getenv("TABLE_NAME"): requestItems,
		},
	}

	resp, err := app.config.db.BatchWriteItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}

	if len(resp.UnprocessedItems) != 0 {
		// NOTE(SMT): need to implement with exponential backup
		log.Printf("[ERROR] IDs %v not deleted, retrying", resp.UnprocessedItems)
	}

	return nil
}

func (app application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := deleteRepositoriesEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	err = app.stageBatchWrites(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to delete repos %v, %v", e, err), 404, err, headers, []string{})
		return resp, nil
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Deleted repos %v successfully", e), 200, err, headers, []string{})
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
