package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/modules"
)

type deleteUserEvent struct {
	EmailAddress string `json:"email_address"`
}

type response events.APIGatewayProxyResponse

var client *cidp.CognitoIdentityProvider

func init() {
	client = cidp.New(session.New())
}

func deleteUser(e deleteUserEvent) error {
	input := &cidp.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
		Username:   aws.String(e.EmailAddress),
	}
	_, err := client.AdminDeleteUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cidp.ErrCodeResourceNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeResourceNotFoundException, aerr.Error())
			case cidp.ErrCodeInvalidParameterException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidParameterException, aerr.Error())
			case cidp.ErrCodeTooManyRequestsException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeTooManyRequestsException, aerr.Error())
			case cidp.ErrCodeNotAuthorizedException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeNotAuthorizedException, aerr.Error())
			case cidp.ErrCodeUserNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserNotFoundException, aerr.Error())
			case cidp.ErrCodeInternalErrorException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInternalErrorException, aerr.Error())
			default:
				log.Printf("[ERROR] %v", err.Error())
			}
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Deleted user %v successfully", e.EmailAddress)
	return nil
}

func handler(ctx context.Context, e deleteUserEvent) (response, error) {
	err := deleteUser(e)
	var body string
	var buf bytes.Buffer
	var statusCode int

	if err != nil {
		statusCode = 404
		body = fmt.Sprintf("Error deleting user %v, %v", e.EmailAddress, err.Error())
	} else {
		statusCode = 200
		body = fmt.Sprintf("Deleted user %v", e.EmailAddress)
	}

	buf, statusCode = modules.GenerateResponseBody(body, statusCode)

	resp := response{
		StatusCode:      statusCode,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
