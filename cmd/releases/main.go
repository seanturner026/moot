package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/google/go-github/github"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

// releaseEvent is an API Gateway POST which contains information necessary to create a release on
// github.com or gitlab.com
type releaseEvent struct {
	RepoOwner       string `json:"repo_owner"`
	RepoName        string `json:"repo_name"`
	RepoProvider    string `json:"repo_provider"`
	BranchBase      string `json:"branch_base"`
	BranchHead      string `json:"branch_head"`
	ReleaseBody     string `json:"release_body"`
	ReleaseVersion  string `json:"release_version"`
	GitlabProjectID string `json:"gitlab_project_id,omitempty"`
	Hotfix          bool   `json:"hotfix"`
}

type application struct {
	aws    awsController
	gh     githubController
	gl     gitlabController
	Config configuration
}

type awsController struct {
	TableName string
	db        dynamodbiface.DynamoDBAPI
}

type configuration struct {
	SlackWebhookURL string
}

func (app awsController) updateCurrentVersion(e releaseEvent) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":cv": {
				S: aws.String(e.ReleaseVersion),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("repo"),
			},
			"SK": {
				S: aws.String(fmt.Sprintf("%s#%s", e.RepoProvider, e.RepoName)),
			},
		},
		TableName:        aws.String(app.TableName),
		UpdateExpression: aws.String("SET CurrentVersion = :cv"),
	}

	log.Printf("[INFO] updating %v latest version to %v...", e.RepoName, e.ReleaseVersion)
	_, err := app.db.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	return nil
}

// handler executes the release and notification workflow
func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	e := releaseEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	var message string
	var statusCode int
	if event.RawPath == "/releases/create/github" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode = app.releasesGithubHandler(e)

	} else if event.RawPath == "/releases/create/gitlab" {
		log.Printf("[INFO] handling request on %v", event.RawPath)
		message, statusCode = app.releasesGitlabHandler(e)

	} else {
		log.Printf("[ERROR] path %v does not exist", event.RawPath)
		return util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{}), nil
	}

	if app.Config.SlackWebhookURL != "" {
		err = util.PostToSlack(app.Config.SlackWebhookURL, fmt.Sprintf(
			"Starting release for %v version %v...\n\n%v",
			e.RepoName,
			e.ReleaseVersion,
			e.ReleaseBody,
		))
		if err != nil {
			message := fmt.Sprintf("Released %v version %v successfully, unable to send slack notification and update latest version in backend", e.RepoName, e.ReleaseVersion)
			statusCode := 200
			return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
		}
	}

	err = app.aws.updateCurrentVersion(e)
	if err != nil {
		message := fmt.Sprintf("Released %v version %v successfully, unable to update latest version in backend", e.RepoName, e.ReleaseVersion)
		statusCode := 200
		return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
	}

	return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
}

func main() {
	app := application{
		aws: awsController{
			TableName: os.Getenv("TABLE_NAME"),
			db:        dynamodb.New(session.Must(session.NewSession())),
		},
		Config: configuration{
			SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
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
			ProjectID:          "",
			MergeRequestSquash: false,
			RemoveSourceBranch: true,
			Client:             clientGitlab,
		}
	}

	lambda.Start(app.handler)
}
