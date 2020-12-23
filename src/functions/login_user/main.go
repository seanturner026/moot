package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	lib "github.com/seanturner026/serverless-release-dashboard/lib"
)

type loginUserEvent struct {
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
}

type response events.APIGatewayProxyResponse

var client *cidp.CognitoIdentityProvider

func init() {
	client = cidp.New(session.New())
}

func getUserPoolClientSecret() (string, error) {
	input := &cidp.DescribeUserPoolClientInput{
		UserPoolId: aws.String(os.Getenv("USER_POOL_ID")),
		ClientId:   aws.String(os.Getenv("CLIENT_POOL_ID")),
	}

	resp, err := client.DescribeUserPoolClient(input)
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
			case cidp.ErrCodeInternalErrorException:
				log.Printf("[ERROR] %v, %v", cidp.ErrCodeInternalErrorException, aerr.Error())
			default:
				log.Printf("[ERROR] %v", err.Error())
			}
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}
	log.Println("[INFO] Obtained user pool client secret successfully")
	return *resp.UserPoolClient.ClientSecret, nil
}

func generateSecretHash(e loginUserEvent, clientSecret string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(e.EmailAddress + os.Getenv("CLIENT_POOL_ID")))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func loginUser(e loginUserEvent, secretHash string) (string, string, bool, error) {
	input := &cidp.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME":    aws.String(e.EmailAddress),
			"PASSWORD":    aws.String(e.Password),
			"SECRET_HASH": aws.String(secretHash),
		},
		ClientId: aws.String(os.Getenv("CLIENT_POOL_ID")),
	}

	resp, err := client.InitiateAuth(input)
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
		return "", "", false, err
	}

	if resp.ChallengeName == aws.String("NEW_PASSWORD_REQUIRED") {
		log.Printf("[INFO] New password required for %v", e.EmailAddress)
		return "", *resp.Session, true, nil
	}

	log.Printf("[INFO] Authenticated user %v successfully", e.EmailAddress)
	return *resp.AuthenticationResult.AccessToken, "", false, nil
}

func handler(ctx context.Context, e loginUserEvent) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	clientSecret, err := getUserPoolClientSecret()
	if err != nil {
		resp := lib.GenerateResponseBody("Error obtaining user pool client secret", 404, err, headers)
		return resp, nil
	}

	secretHash := generateSecretHash(e, clientSecret)
	accessToken, session, resetPassword, err := loginUser(e, secretHash)
	if err != nil {
		resp := lib.GenerateResponseBody(fmt.Sprintf("Error logging user %v in", e.EmailAddress), 404, err, headers)
		return resp, nil

	} else if resetPassword {
		headers["X-Session-Id"] = session
		resp := lib.GenerateResponseBody(
			fmt.Sprintf("User %v logged in successfully, password change required", e.EmailAddress), 200, err, headers,
		)
		return resp, nil
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %v", accessToken)
	resp := lib.GenerateResponseBody(fmt.Sprintf("User %v logged in successfully", e.EmailAddress), 200, err, headers)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
