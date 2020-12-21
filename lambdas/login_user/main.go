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

type loginUserEvent struct {
	EmailAddress string
	Password     string
}

type response events.APIGatewayProxyResponse

var client *cidp.CognitoIdentityProvider

func init() {
	client = cidp.New(session.New())
}

// NOTE(SMT): Need to use InitiateAuth() and not AdminInitiateAuth()?
func loginUser(e loginUserEvent) error {
	input := &cidp.AdminInitiateAuthInput{
		AuthFlow: aws.String("ADMIN_NO_SRP_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(e.EmailAddress),
			"PASSWORD": aws.String(e.Password),
		},
		ClientId:   aws.String(os.Getenv("CLIENT_POOL_ID")),
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
	}

	resp, err := client.AdminInitiateAuth(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cidp.ErrCodeResourceNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeResourceNotFoundException, aerr.Error())
			case cidp.ErrCodeInvalidParameterException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidParameterException, aerr.Error())
			case cidp.ErrCodeNotAuthorizedException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeNotAuthorizedException, aerr.Error())
			case cidp.ErrCodeTooManyRequestsException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeTooManyRequestsException, aerr.Error())
			case cidp.ErrCodeInternalErrorException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInternalErrorException, aerr.Error())
			case cidp.ErrCodeUnexpectedLambdaException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUnexpectedLambdaException, aerr.Error())
			case cidp.ErrCodeInvalidUserPoolConfigurationException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidUserPoolConfigurationException, aerr.Error())
			case cidp.ErrCodeUserLambdaValidationException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserLambdaValidationException, aerr.Error())
			case cidp.ErrCodeInvalidLambdaResponseException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidLambdaResponseException, aerr.Error())
			case cidp.ErrCodeMFAMethodNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeMFAMethodNotFoundException, aerr.Error())
			case cidp.ErrCodeInvalidSmsRoleAccessPolicyException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidSmsRoleAccessPolicyException, aerr.Error())
			case cidp.ErrCodeInvalidSmsRoleTrustRelationshipException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidSmsRoleTrustRelationshipException, aerr.Error())
			case cidp.ErrCodePasswordResetRequiredException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodePasswordResetRequiredException, aerr.Error())
			case cidp.ErrCodeUserNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserNotFoundException, aerr.Error())
			case cidp.ErrCodeUserNotConfirmedException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserNotConfirmedException, aerr.Error())
			default:
				log.Printf("[ERROR] %v", err.Error())
			}
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}

	log.Printf("[INFO] Authenticated user %v successfully", e.EmailAddress)
	if resp.ChallengeName == aws.String("NEW_PASSWORD_REQUIRED") {
		return nil
	}

	return nil
}

func handler(ctx context.Context, e loginUserEvent) (response, error) {
	err := loginUser(e)
	var body string
	var buf bytes.Buffer
	var statusCode int

	if err != nil {
		statusCode = 404
		body = fmt.Sprintf("Error logging user %v in, %v", e.EmailAddress, err.Error())
	} else {
		statusCode = 200
		body = fmt.Sprintf("User %v logged in successfully", e.EmailAddress)
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
