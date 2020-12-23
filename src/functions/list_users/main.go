package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	lib "github.com/seanturner026/serverless-release-dashboard/lib"
)

var client *cidp.CognitoIdentityProvider

func init() {
	client = cidp.New(session.New())
}

func listUsers() ([]string, error) {
	input := &cidp.ListUsersInput{
		AttributesToGet: aws.StringSlice([]string{"email"}),
		Limit:           aws.Int64(60),
		UserPoolId:      aws.String(os.Getenv("USER_POOL_ID")),
	}

	resp, err := client.ListUsers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cidp.ErrCodeInvalidParameterException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidParameterException, aerr.Error())
			case cidp.ErrCodeResourceNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeResourceNotFoundException, aerr.Error())
			case cidp.ErrCodeTooManyRequestsException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeTooManyRequestsException, aerr.Error())
			case cidp.ErrCodeNotAuthorizedException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeNotAuthorizedException, aerr.Error())
			case cidp.ErrCodeInternalErrorException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInternalErrorException, aerr.Error())
			default:
				log.Printf("[ERROR] %v", err.Error())
			}
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return []string{}, err
	}

	// NOTE(SMT): reflect []*cidp.UserType to stringslice?
	users := resp.Users
	userNames := make([]string, len(users))
	for i, user := range users {
		if i == 0 {
			userNames[i] = *user.Attributes[0].Value
		} else {
			userNames = append(userNames, *user.Attributes[0].Value)
		}
	}

	return userNames, nil
}

func handler(ctx context.Context, e interface{}) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	userNames, err := listUsers()
	if err != nil || len(userNames) == 0 {
		resp := lib.GenerateResponseBody("Unable to populate list of users", 404, err, headers)
		return resp, nil
	}

	resp := lib.GenerateResponseBody(strings.Join(userNames, ","), 200, err, headers)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
