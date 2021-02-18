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
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type createRepoEvent struct {
	RepoName     string `dynamodbav:"PK" json:"repo_name"`
	RepoProvider string `dynamodbav:"RepoProvider" json:"repo_provider"`
	RepoOwner    string `dynamodbav:"SK" json:"repo_owner"`
	BranchBase   string `dynamodbav:"BranchBase" json:"branch_base"`
	BranchHead   string `dynamodbav:"BranchHead" json:"branch_head"`
}

func generatePutItemInput(e createRepoEvent) (createRepoEvent, map[string]*dynamodb.AttributeValue, error) {
	e.RepoOwner = fmt.Sprintf("repo#%v", e.RepoOwner)
	itemInput, err := dynamodbattribute.MarshalMap(e)
	log.Printf("[DEBUG] itemInput %+v", itemInput)
	if err != nil {
		return e, map[string]*dynamodb.AttributeValue{}, err
	}
	return e, itemInput, nil
}

func (app application) writeRepoToDB(e createRepoEvent, itemInput map[string]*dynamodb.AttributeValue) error {
	input := &dynamodb.PutItemInput{
		ReturnConsumedCapacity: aws.String("TOTAL"),
		TableName:              aws.String(app.config.TableName),
		Item:                   itemInput,
	}
	_, err := app.config.db.PutItem(input)
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

func (app application) repositoriesCreateHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) events.APIGatewayV2HTTPResponse {

	e := createRepoEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	e, itemInput, err := generatePutItemInput(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to stage provided information for loading into DynamoDB for ID %v, %v", e.RepoName, err), 404, err, headers, []string{})
		return resp
	}

	err = app.writeRepoToDB(e, itemInput)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Failed to write record %v to DynamoDB table, %v", e.RepoName, err), 404, err, headers, []string{})
		return resp
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Wrote record %v to DynamoDB successfully", e.RepoName), 200, nil, headers, []string{})
	return resp
}
