package main

import (
	"encoding/json"
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
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type resetPasswordEvent struct {
	EmailAddress string `json:"email_address"`
	NewPassword  string `json:"new_password"`
	SessionID    string `json:"session_id"`
}

type application struct {
	config configuration
}

type configuration struct {
	ClientPoolID string
	UserPoolID   string
	idp          cidpif.CognitoIdentityProviderAPI
}

func (app *application) getUserPoolClientSecret() (string, error) {
	input := &cidp.DescribeUserPoolClientInput{
		UserPoolId: aws.String(app.config.UserPoolID),
		ClientId:   aws.String(app.config.ClientPoolID),
	}

	resp, err := app.config.idp.DescribeUserPoolClient(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}
	log.Println("[INFO] Obtained user pool client secret successfully")
	return *resp.UserPoolClient.ClientSecret, nil
}

func (app *application) resetPassword(e resetPasswordEvent, secretHash string) (string, error) {
	input := &cidp.AdminRespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(e.EmailAddress),
			"NEW_PASSWORD": aws.String(e.NewPassword),
			"SECRET_HASH":  aws.String(secretHash),
		},
		ClientId:   aws.String(app.config.ClientPoolID),
		UserPoolId: aws.String(app.config.UserPoolID),
		Session:    aws.String(e.SessionID),
	}

	resp, err := app.config.idp.AdminRespondToAuthChallenge(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}

	return *resp.AuthenticationResult.AccessToken, nil
}

func (app *application) handler(event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := resetPasswordEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	clientSecret, err := app.getUserPoolClientSecret()
	if err != nil {
		resp := util.GenerateResponseBody("Error obtaining user pool client secret", 404, err, headers)
		return resp, nil
	}

	secretHash := util.GenerateSecretHash(clientSecret, e.EmailAddress, app.config.ClientPoolID)
	AccessToken, err := app.resetPassword(e, secretHash)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error changing user %v password", e.EmailAddress), 404, err, headers)
		return resp, nil
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %v", AccessToken)
	resp := util.GenerateResponseBody(fmt.Sprintf("User %v changed password successfully", e.EmailAddress), 200, err, headers)
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
