package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockPutItem struct {
	dynamodbiface.DynamoDBAPI
	Response *dynamodb.PutItemOutput
	Error    error
}

func (m mockPutItem) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return m.Response, nil
}

func TestGeneratePutItemInupt(t *testing.T) {
	t.Run("Successfully marshalled json for DynamoDB", func(t *testing.T) {
		event := createRepoEvent{
			RepoName:   "test",
			RepoOwner:  "test",
			BranchHead: "test",
			BranchBase: "test",
		}

		_, err := generatePutItemInputExpression(event)
		if err != nil {
			t.Fatal("Input should have been marshalled for DynamoDB")
		}
	})
}

func TestWriteRepoToDB(t *testing.T) {
	t.Run("Successfully create repo in DynamoDB", func(t *testing.T) {
		dbMock := mockPutItem{
			Response: &dynamodb.PutItemOutput{},
			Error:    nil,
		}

		app := application{AWS: awsController{
			TableName: "test",
			DB:        dbMock,
		}}

		event := createRepoEvent{
			RepoName:   "test",
			RepoOwner:  "test",
			BranchHead: "test",
			BranchBase: "test",
		}

		input, _ := generatePutItemInputExpression(event)

		err := app.AWS.writeRepoToDB(event, input)
		if err != nil {
			t.Fatal("Input should have been written to DynamoDB")
		}
	})
}
