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
	"github.com/google/go-github/github"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type application struct {
	aws awsController
	gh  githubController
	gl  gitlabController
}

type awsController struct {
	GlobalSecondaryIndexName string
	TableName                string
	db                       dynamodbiface.DynamoDBAPI
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
	var resp events.APIGatewayV2HTTPResponse
	headers := map[string]string{"Content-Type": "application/json"}

	if event.RawPath == "/repositories/create" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.repositoriesCreateHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/delete" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.repositoriesDeleteHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else if event.RawPath == "/repositories/list" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode := app.repositoriesListHandler(event)
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		resp = util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{})
		return resp, nil
	}
}

func main() {
	app := application{
		aws: awsController{
			GlobalSecondaryIndexName: os.Getenv("GLOBAL_SECONDARY_INDEX_NAME"),
			TableName:                os.Getenv("TABLE_NAME"),
			db:                       dynamodb.New(session.Must(session.NewSession())),
		},
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		githubCtx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
		tc := oauth2.NewClient(githubCtx, ts)

		app.gh = githubController{
			Client:    github.NewClient(tc),
			GithubCtx: githubCtx,
		}
	}

	if os.Getenv("GITLAB_TOKEN") != "" {
		clientGitlab, err := gitlab.NewClient(os.Getenv("GITLAB_TOKEN"))
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
		app.gl = gitlabController{
			MergeRequestSquash: false,
			RemoveSourceBranch: true,
			Client:             clientGitlab,
		}
	}

	lambda.Start(app.handler)
}
