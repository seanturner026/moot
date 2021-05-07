package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
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
	AWS    awsController
	GH     githubController
	GL     gitlabController
	Config configuration
}

type awsController struct {
	TableName string
	DB        dynamodbiface.DynamoDBAPI
	SSM       ssmiface.SSMAPI
}

type configuration struct {
	SlackWebhookURL string
}

func (app awsController) getProviderToken(e releaseEvent) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           aws.String(fmt.Sprintf("/dev_release_dashboard/%s_token", e.RepoProvider)),
		WithDecryption: aws.Bool(true),
	}

	resp, err := app.SSM.GetParameter(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return "", err
	}

	token := *resp.Parameter.Value
	return token, nil
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

	log.Info(fmt.Sprintf("updating %v latest version to %v...", e.RepoName, e.ReleaseVersion))
	_, err := app.DB.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
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
		log.Error(fmt.Sprintf("%v", err))
	}

	token, err := app.AWS.getProviderToken(e)
	if err != nil {
		message := fmt.Sprintf("Unable to release %s version %s, please double check the %s token", e.RepoName, e.ReleaseVersion, e.RepoProvider)
		statusCode := 400
		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
	}

	var message string
	var statusCode int
	if event.RawPath == "/releases/create/github" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode = app.releasesGithubHandler(e, token)

	} else if event.RawPath == "/releases/create/gitlab" {
		log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
		message, statusCode = app.releasesGitlabHandler(e, token)

	} else {
		log.Error(fmt.Sprintf("path %v does not exist", event.RawPath))
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

	err = app.AWS.updateCurrentVersion(e)
	if err != nil {
		message := fmt.Sprintf("Released %v version %v successfully, unable to update latest version in backend", e.RepoName, e.ReleaseVersion)
		statusCode := 200
		return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
	}

	return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	app := application{
		AWS: awsController{
			TableName: os.Getenv("TABLE_NAME"),
			DB:        dynamodb.New(session.Must(session.NewSession())),
			SSM:       ssm.New(session.Must(session.NewSession())),
		},
		Config: configuration{
			SlackWebhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		},
	}

	lambda.Start(app.handler)
}
