package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockQuery struct {
	dynamodbiface.DynamoDBAPI
	Response *dynamodb.QueryOutput
	Error    error
}

func (m mockQuery) Query(*dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return m.Response, nil
}

func TestListRepos(t *testing.T) {
	t.Run("Sucessfully queried DynamoDB for items", func(t *testing.T) {
		dbMock := mockQuery{
			Response: &dynamodb.QueryOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			TableName: "test",
			db:        dbMock,
		}}

		event := listReposEvent{RepoOwner: "test"}

		_, err := app.listRepos(event)
		if err != nil {
			t.Fatal("Query should have returned results")
		}
	})
}
