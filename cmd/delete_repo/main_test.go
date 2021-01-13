package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDeleteItem struct {
	dynamodbiface.DynamoDBAPI
	Response *dynamodb.DeleteItemOutput
	Error    error
}

func (m mockDeleteItem) DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return m.Response, nil
}

func TestDeleteRepo(t *testing.T) {
	t.Run("Successfully delete repo in DynamoDB", func(t *testing.T) {
		dbMock := mockDeleteItem{
			Response: &dynamodb.DeleteItemOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			TableName: "test",
			db:        dbMock,
		}}

		event := deleteRepoEvent{RepoName: "test"}

		err := app.deleteRepo(event)
		if err != nil {
			t.Fatal("Repo should have been deleted")
		}
	})
}
