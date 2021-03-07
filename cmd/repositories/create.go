package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/xanzy/go-gitlab"
)

type createRepoEvent struct {
	Type            string `dynamodbav:"PK" json:"type"`
	RepoProvider    string `dynamodbav:"SK" json:"repo_provider"`
	RepoName        string `dynamodbav:"-" json:"repo_name"`
	RepoOwner       string `dynamodbav:"RepoOwner" json:"repo_owner"`
	BranchBase      string `dynamodbav:"BranchBase" json:"branch_base"`
	BranchHead      string `dynamodbav:"BranchHead" json:"branch_head"`
	GitlabProjectID string `dynamodbav:"GitlabProjectID,omitempty" json:"gitlab_repo_id,omitempty"`
}

func (app githubController) confirmTokenAccess(e createRepoEvent) error {
	_, _, err := app.Client.Repositories.Get(app.GithubCtx, e.RepoOwner, e.RepoName)
	if err != nil {
		return err
	}
	return nil
}

func (app gitlabController) confirmTokenAccess(e createRepoEvent) error {
	_, _, err := app.Client.Projects.GetProject(e.GitlabProjectID, &gitlab.GetProjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func generatePutItemInput(e createRepoEvent) (createRepoEvent, map[string]*dynamodb.AttributeValue, error) {
	e.RepoProvider = fmt.Sprintf("%s#%s", e.RepoProvider, e.RepoName)
	itemInput, err := dynamodbattribute.MarshalMap(e)
	if err != nil {
		return e, map[string]*dynamodb.AttributeValue{}, err
	}
	return e, itemInput, nil
}

func (app awsController) writeRepoToDB(e createRepoEvent, itemInput map[string]*dynamodb.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(app.TableName),
		Item:                   itemInput,
	}
	_, err := app.db.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Wrote ID %s successfully", e.RepoName)
	return nil
}

func (app application) repositoriesCreateHandler(event events.APIGatewayV2HTTPRequest) (string, int) {
	e := createRepoEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	if e.RepoProvider == "github.com" {
		err = app.gh.confirmTokenAccess(e)
	} else if e.RepoProvider == "gitlab.com" {
		err = app.gl.confirmTokenAccess(e)
	}

	if err != nil {
		message := fmt.Sprintf("Provided %s token is unable to access repository %s", e.RepoProvider, e.RepoName)
		statusCode := 401
		return message, statusCode
	}

	e, itemInput, err := generatePutItemInput(e)
	if err != nil {
		message := fmt.Sprintf("Failed to stage provided information for loading into DynamoDB for ID %s", e.RepoName)
		statusCode := 400
		return message, statusCode
	}

	err = app.aws.writeRepoToDB(e, itemInput)
	if err != nil {
		message := fmt.Sprintf("Failed to write record %s to DynamoDB table", e.RepoName)
		statusCode := 400
		return message, statusCode
	}

	message := fmt.Sprintf("Wrote record %s to DynamoDB successfully", e.RepoName)
	statusCode := 200
	return message, statusCode
}
