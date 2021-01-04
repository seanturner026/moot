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

type loginUserEvent struct {
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
}

type application struct {
	config configuration
}

type configuration struct {
	ClientPoolID string
	UserPoolID   string
	idp          cidpif.CognitoIdentityProviderAPI
}

type loginUserResponse struct {
	AccessToken         string `json:"access_token,omitempty"`
	NewPasswordRequired bool
	SessionID           string `json:"session_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
}

func (app *application) getUserPoolClientSecret() (string, error) {
	input := &cidp.DescribeUserPoolClientInput{
		UserPoolId: aws.String(app.config.UserPoolID),
		ClientId:   aws.String(app.config.ClientPoolID),
	}

	resp, err := app.config.idp.DescribeUserPoolClient(input)
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

func (app *application) loginUser(e loginUserEvent, secretHash string) (loginUserResponse, error) {
	input := &cidp.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME":    aws.String(e.EmailAddress),
			"PASSWORD":    aws.String(e.Password),
			"SECRET_HASH": aws.String(secretHash),
		},
		ClientId: aws.String(os.Getenv("CLIENT_POOL_ID")),
	}

	loginUserResp := loginUserResponse{}
	resp, err := app.config.idp.InitiateAuth(input)
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
		loginUserResp.NewPasswordRequired = false
		return loginUserResp, err
	}

	if *resp.ChallengeName == "NEW_PASSWORD_REQUIRED" {
		log.Printf("[INFO] New password required for %v", e.EmailAddress)
		loginUserResp.NewPasswordRequired = true
		loginUserResp.SessionID = *resp.Session
		loginUserResp.UserID = *resp.ChallengeParameters["USER_ID_FOR_SRP"]
		return loginUserResp, nil
	}

	log.Printf("[INFO] Authenticated user %v successfully", e.EmailAddress)
	loginUserResp.AccessToken = *resp.AuthenticationResult.AccessToken
	loginUserResp.NewPasswordRequired = false
	return loginUserResp, nil
}

func (app *application) handler(e loginUserEvent) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	clientSecret, err := app.getUserPoolClientSecret()
	if err != nil {
		resp := util.GenerateResponseBody("Error obtaining user pool client secret", 404, err, headers)
		return resp, nil
	}

	secretHash := util.GenerateSecretHash(clientSecret, e.EmailAddress, app.config.ClientPoolID)
	loginUserResp, err := app.loginUser(e, secretHash)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error logging user %v in", e.EmailAddress), 404, err, headers)
		return resp, nil

	} else if loginUserResp.NewPasswordRequired {
		headers["X-Session-Id"] = loginUserResp.SessionID
		headers["X-User-Id"] = loginUserResp.UserID
		resp := util.GenerateResponseBody(
			fmt.Sprintf("User %v logged in successfully, password change required", e.EmailAddress), 200, err, headers,
		)
		return resp, nil
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %v", loginUserResp.AccessToken)
	resp := util.GenerateResponseBody(fmt.Sprintf("User %v logged in successfully", e.EmailAddress), 200, err, headers)
	return resp, nil
}

func main() {
	config := configuration{
		ClientPoolID: os.Getenv("CLIENT_POOL_ID"),
		UserPoolID:   os.Getenv("USER_POOL_ID"),
		idp:          cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
