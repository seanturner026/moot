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

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

type response events.APIGatewayProxyResponse

var client *cidp.CognitoIdentityProvider

func init() {
	client = cidp.New(session.New())
}

func createUser(e createUserEvent) error {
	input := &cidp.AdminCreateUserInput{
		UserPoolId:             aws.String(os.Getenv("USER_POOL_ID")),
		Username:               aws.String(e.EmailAddress),
		DesiredDeliveryMediums: []*string{aws.String("EMAIL")},
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cidp.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.EmailAddress),
			},
		},
	}
	_, err := client.AdminCreateUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case cidp.ErrCodeResourceNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeResourceNotFoundException, aerr.Error())
			case cidp.ErrCodeInvalidParameterException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidParameterException, aerr.Error())
			case cidp.ErrCodeUserNotFoundException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserNotFoundException, aerr.Error())
			case cidp.ErrCodeUsernameExistsException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUsernameExistsException, aerr.Error())
			case cidp.ErrCodeInvalidPasswordException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidPasswordException, aerr.Error())
			case cidp.ErrCodeCodeDeliveryFailureException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeCodeDeliveryFailureException, aerr.Error())
			case cidp.ErrCodeUnexpectedLambdaException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUnexpectedLambdaException, aerr.Error())
			case cidp.ErrCodeUserLambdaValidationException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUserLambdaValidationException, aerr.Error())
			case cidp.ErrCodeInvalidLambdaResponseException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidLambdaResponseException, aerr.Error())
			case cidp.ErrCodePreconditionNotMetException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodePreconditionNotMetException, aerr.Error())
			case cidp.ErrCodeInvalidSmsRoleAccessPolicyException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidSmsRoleAccessPolicyException, aerr.Error())
			case cidp.ErrCodeInvalidSmsRoleTrustRelationshipException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInvalidSmsRoleTrustRelationshipException, aerr.Error())
			case cidp.ErrCodeTooManyRequestsException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeTooManyRequestsException, aerr.Error())
			case cidp.ErrCodeNotAuthorizedException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeNotAuthorizedException, aerr.Error())
			case cidp.ErrCodeUnsupportedUserStateException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeUnsupportedUserStateException, aerr.Error())
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
	log.Printf("[INFO] Created new user %v successfully", e.EmailAddress)
	return nil
}

func handler(ctx context.Context, e createUserEvent) (response, error) {
	err := createUser(e)
	var body string
	var buf bytes.Buffer
	var statusCode int

	if err != nil {
		statusCode = 404
		body = fmt.Sprintf("Error creating user %v, %v", e.EmailAddress, err.Error())
	} else {
		statusCode = 200
		body = fmt.Sprintf("Created new user %v", e.EmailAddress)
	}

	buf, statusCode = modules.GenerateResponseBody(body, statusCode)
	message := fmt.Sprintf(`Successfully created user %v, please have the user run the following command to obtain access:

	aws cognito-idp admin-set-user-password \
		--user-pool-id %v \
		--username %v \
		--password <GENERATE_PASSWORD> \
		--permanent \
		--region %v
	`, e.EmailAddress, os.Getenv("USER_POOL_ID"), e.EmailAddress, os.Getenv("REGION"))

	modules.PostToSlack(os.Getenv("WEBHOOK_URL"), message)

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
