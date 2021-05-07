package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
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
	log.Info("deleted IDs successfully")
	return nil
}

func (app awsController) deleteRepositories(requestItems []*dynamodb.WriteRequest) error {
	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			os.Getenv("TABLE_NAME"): requestItems,
		},
	}

	resp, err := app.DB.BatchWriteItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return err
	}

	if len(resp.UnprocessedItems) != 0 {
		// NOTE(SMT): need to implement
		log.Error(fmt.Sprintf("ids %v not deleted, retrying", resp.UnprocessedItems))
	}

	return nil
}

func (app application) repositoriesDeleteHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := deleteRepositoriesEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	err = app.AWS.stageBatchWrites(e)
	if err != nil {
		message := "Failed to delete repositories"
		statusCode := 400
		return message, statusCode
	}

	message := "Deleted repositories successfully"
	statusCode := 200
	return message, statusCode
}
