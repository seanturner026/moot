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

func TestListUsers(t *testing.T) {
	t.Run("Successfully listed users", func(t *testing.T) {
		dbMock := mockQuery{
			Response: &dynamodb.QueryOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			TableName: "test",
			DB:        dbMock,
		}}
		tenantID := "1234567890"

		_, err := app.listUsers(tenantID)
		if err != nil {
			t.Fatal("Users should have been listed")
		}
	})
}
