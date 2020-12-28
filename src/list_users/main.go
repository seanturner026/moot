package main

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type application struct {
	config configuration
}

type configuration struct {
	UserPoolID string
	idp        cidpif.CognitoIdentityProviderAPI
}

func (app *application) listUsers() ([]string, error) {
	input := &cidp.ListUsersInput{
		AttributesToGet: aws.StringSlice([]string{"email"}),
		Limit:           aws.Int64(60),
		UserPoolId:      aws.String(app.config.UserPoolID),
	}

	resp, err := app.config.idp.ListUsers(input)
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

func (app *application) handler() (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	userNames, err := app.listUsers()
	if err != nil || len(userNames) == 0 {
		resp := util.GenerateResponseBody("Unable to populate list of users", 404, err, headers)
		return resp, nil
	}

	resp := util.GenerateResponseBody(strings.Join(userNames, ","), 200, err, headers)
	return resp, nil
}

func main() {
	config := configuration{
		UserPoolID: os.Getenv("USER_POOL_ID"),
		idp:        cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
