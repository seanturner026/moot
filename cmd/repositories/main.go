package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type application struct {
	config configuration
}

type configuration struct {
	GlobalSecondaryIndexName string
	TableName                string
	db                       dynamodbiface.DynamoDBAPI
}

type repository struct {
	RepoName        string `json:"repo_name,omitempty"`
	RepoProvider    string `dynamodbav:"SK" json:"repo_provider,omitempty"`
	RepoOwner       string `dynamodbav:"RepoOwner" json:"repo_owner,omitempty"`
	BranchBase      string `dynamodbav:"BranchBase" json:"branch_base,omitempty"`
	BranchHead      string `dynamodbav:"BranchHead" json:"branch_head,omitempty"`
	CurrentVersion  string `dynamodbav:"CurrentVersion" json:"current_version,omitempty"`
	GitlabProjectID string `dynamodbav:"GitlabProjectID,omitempty" json:"gitlab_repo_id,omitempty"`
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/repositories/create" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.repositoriesCreateHandler(event, headers)
		return resp, nil

	} else if event.RawPath == "/repositories/delete" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.repositoriesDeleteHandler(event, headers)
		return resp, nil

	} else if event.RawPath == "/repositories/list" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		resp = app.repositoriesListHandler(event, headers)
		return resp, nil

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
		return resp, nil
	}
}

func main() {
	config := configuration{
		GlobalSecondaryIndexName: os.Getenv("GLOBAL_SECONDARY_INDEX_NAME"),
		TableName:                os.Getenv("TABLE_NAME"),
		db:                       dynamodb.New(session.Must(session.NewSession())),
	}

	app := application{config: config}
	lambda.Start(app.handler)
}
