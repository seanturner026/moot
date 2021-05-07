package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
)

type application struct {
	Config configuration
}

type configuration struct {
	DB         dynamodbiface.DynamoDBAPI
	TableName  string
	UserPoolID string
}

func (app application) handler(event events.CognitoEventUserPoolsPostConfirmation) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	log.Info(fmt.Sprintf("%+v", event))

	// e := releaseEvent{}
	// err := json.Unmarshal([]byte(event), &e)
	// if err != nil {
	// 	log.Error(fmt.Sprintf("%v", err))
	// }

	// token, err := app.AWS.getProviderToken(e)
	// if err != nil {
	// 	message := fmt.Sprintf("Unable to release %s version %s, please double check the %s token", e.RepoName, e.ReleaseVersion, e.RepoProvider)
	// 	statusCode := 400
	// 	return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
	// }

	// var message string
	// var statusCode int
	// if event.RawPath == "/releases/create/github" {
	// 	log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
	// 	message, statusCode = app.releasesGithubHandler(e, token)

	// } else if event.RawPath == "/releases/create/gitlab" {
	// 	log.Info(fmt.Sprintf("handling request on %v", event.RawPath))
	// 	message, statusCode = app.releasesGitlabHandler(e, token)

	// } else {
	// 	log.Error(fmt.Sprintf("path %v does not exist", event.RawPath))
	// 	return util.GenerateResponseBody(fmt.Sprintf("Path does not exist %v", event.RawPath), 404, nil, headers, []string{}), nil
	// }

	// if app.Config.SlackWebhookURL != "" {
	// 	err = util.PostToSlack(app.Config.SlackWebhookURL, fmt.Sprintf(
	// 		"Starting release for %v version %v...\n\n%v",
	// 		e.RepoName,
	// 		e.ReleaseVersion,
	// 		e.ReleaseBody,
	// 	))
	// 	if err != nil {
	// 		message := fmt.Sprintf("Released %v version %v successfully, unable to send slack notification and update latest version in backend", e.RepoName, e.ReleaseVersion)
	// 		statusCode := 200
	// 		return util.GenerateResponseBody(message, statusCode, nil, headers, []string{}), nil
	// 	}
	// }

	// err = app.AWS.updateCurrentVersion(e)
	// if err != nil {
	// 	message := fmt.Sprintf("Released %v version %v successfully, unable to update latest version in backend", e.RepoName, e.ReleaseVersion)
	// 	statusCode := 200
	// 	return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
	// }

	return util.GenerateResponseBody("message", 200, nil, headers, []string{}), nil
	// return util.GenerateResponseBody(message, statusCode, err, headers, []string{}), nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := configuration{
		TableName:  os.Getenv("TABLE_NAME"),
		UserPoolID: os.Getenv("USER_POOL_ID"),
		DB:         dynamodb.New(session.Must(session.NewSession())),
	}

	app := application{Config: config}

	lambda.Start(app.handler)
}
