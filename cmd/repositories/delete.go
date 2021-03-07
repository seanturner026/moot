package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type deleteRepositoriesEvent struct {
	Repositories []repository `json:"repositories"`
}

func (app awsController) stageBatchWrites(e deleteRepositoriesEvent) error {
	repositories := []*dynamodb.WriteRequest{}
	for i, r := range e.Repositories {
		if (i+1)%25 == 0 {
			app.deleteRepositories(repositories)
			repositories = []*dynamodb.WriteRequest{}
		}
		deleteRequest := &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: map[string]*dynamodb.AttributeValue{
					"PK": {
						S: aws.String("repo"),
					},
					"SK": {
						S: aws.String(fmt.Sprintf("%s#%s", r.RepoProvider, r.RepoName)),
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

func (app awsController) deleteRepositories(requestItems []*dynamodb.WriteRequest) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			os.Getenv("TABLE_NAME"): requestItems,
		},
	}

	resp, err := app.db.BatchWriteItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}

	if len(resp.UnprocessedItems) != 0 {
		// NOTE(SMT): need to implement
		log.Printf("[ERROR] IDs %v not deleted, retrying", resp.UnprocessedItems)
	}

	return nil
}

func (app application) repositoriesDeleteHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := deleteRepositoriesEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	err = app.aws.stageBatchWrites(e)
	if err != nil {
		message := "Failed to delete repositories"
		statusCode := 400
		return message, statusCode
	}

	message := "Deleted repositories successfully"
	statusCode := 200
	return message, statusCode
}
