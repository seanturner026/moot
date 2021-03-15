package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/google/go-github/github"
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
	"github.com/xanzy/go-gitlab"
)

type application struct {
	aws awsController
	gh  githubController
	gl  gitlabController
}

type awsController struct {
	TableName string
	db        dynamodbiface.DynamoDBAPI
	ssm       ssmiface.SSMAPI
}

type githubController struct {
	Client    *github.Client
	GithubCtx context.Context
}

type gitlabController struct {
	MergeRequestSquash bool
	RemoveSourceBranch bool
	Client             *gitlab.Client
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
	headers := map[string]string{"Content-Type": "application/json"}
	IDToken := event.Headers["x-identity-token"]
	tenantID := util.ExtractTenantID(IDToken)

	if event.RawPath == "/repositories/create" {
		log.Printf("[INFO] handling request on %s", event.RawPath)
		message, statusCode := app.repositoriesCreateHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/delete" {
		log.Printf("[INFO] handling request on %s", event.RawPath)
		message, statusCode := app.repositoriesDeleteHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/list" {
		log.Printf("[INFO] handling request on %s", event.RawPath)
		message, statusCode := app.repositoriesListHandler(event, tenantID)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp := util.GenerateResponseBody(fmt.Sprintf("Path does not exist %s", event.RawPath), 404, nil, headers, []string{})
		return resp, nil
	}
}

func main() {
	app := application{
		aws: awsController{
			TableName: os.Getenv("TABLE_NAME"),
			db:        dynamodb.New(session.Must(session.NewSession())),
			ssm:       ssm.New(session.Must(session.NewSession())),
		},
	}

	lambda.Start(app.handler)
}
