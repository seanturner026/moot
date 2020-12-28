package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type deleteUserEvent struct {
	EmailAddress string `json:"email_address"`
}

type application struct {
	config configuration
}

type configuration struct {
	UserPoolID string
	idp        cidpif.CognitoIdentityProviderAPI
}

func (app *application) deleteUser(e deleteUserEvent) error {
	input := &cidp.AdminDeleteUserInput{
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
		Username:   aws.String(e.EmailAddress),
	}
	_, err := app.config.idp.AdminDeleteUser(input)
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

func (app *application) handler(e deleteUserEvent) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	err := app.deleteUser(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error deleting user %v", e.EmailAddress), 404, err, headers)
		return resp, nil
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Deleted user %v", e.EmailAddress), 200, nil, headers)
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
