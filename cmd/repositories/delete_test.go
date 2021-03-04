package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockBatchWriteItem struct {
	dynamodbiface.DynamoDBAPI
	Response *dynamodb.BatchWriteItemOutput
	Error    error
}

func (m mockBatchWriteItem) BatchWriteItem(*dynamodb.BatchWriteItemInput) (*dynamodb.BatchWriteItemOutput, error) {
	return m.Response, nil
}

func TestStageBatchWrites(t *testing.T) {
	t.Run("Successfully delete stage repos for delete", func(t *testing.T) {
		dbMock := mockBatchWriteItem{
			Response: &dynamodb.BatchWriteItemOutput{},
			Error:    nil,
		}

		app := application{aws: awsController{
			TableName: "test",
			db:        dbMock,
		}}

		event := deleteRepositoriesEvent{Repositories: []repository{
			{RepoName: "test", RepoOwner: "test"},
		}}

		err := app.aws.stageBatchWrites(event)
		if err != nil {
			t.Fatal("Repos should have been staged")
		}
	})
}

func TestDeleteRepositories(t *testing.T) {
	t.Run("Successfully delete repos for in DynamoDB", func(t *testing.T) {
		dbMock := mockBatchWriteItem{
			Response: &dynamodb.BatchWriteItemOutput{},
			Error:    nil,
		}

		app := application{aws: awsController{
			TableName: "test",
			db:        dbMock,
		}}

		requestItems := []*dynamodb.WriteRequest{
			{
				DeleteRequest: &dynamodb.DeleteRequest{
					Key: map[string]*dynamodb.AttributeValue{
						"pk": {
							S: aws.String("test"),
						},
						"sk": {
							S: aws.String("test"),
						},
					},
				},
			},
		}

		err := app.aws.deleteRepositories(requestItems)
		if err != nil {
			t.Fatal("Repo should have been deleted")
		}
	})
}
