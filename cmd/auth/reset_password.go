package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
)

type resetPasswordEvent struct {
	EmailAddress string `json:"email_address"`
	NewPassword  string `json:"new_password"`
	SessionID    string `json:"session_id"`
}

func (app application) resetPassword(e resetPasswordEvent, secretHash string) (string, error) {
	input := &cognitoidentityprovider.AdminRespondToAuthChallengeInput{
		ChallengeName: aws.String("NEW_PASSWORD_REQUIRED"),
		ChallengeResponses: map[string]*string{
			"USERNAME":     aws.String(e.EmailAddress),
			"NEW_PASSWORD": aws.String(e.NewPassword),
			"SECRET_HASH":  aws.String(secretHash),
		},
		ClientId:   aws.String(app.Config.ClientPoolID),
		UserPoolId: aws.String(app.Config.UserPoolID),
		Session:    aws.String(e.SessionID),
	}

	resp, err := app.IDP.AdminRespondToAuthChallenge(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		return "", err
	}

	return *resp.AuthenticationResult.AccessToken, nil
}

func (app application) authResetPasswordHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) (string, int, map[string]string) {
	e := resetPasswordEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	secretHash := util.GenerateSecretHash(app.Config.ClientPoolSecret, e.EmailAddress, app.Config.ClientPoolID)
	AccessToken, err := app.resetPassword(e, secretHash)
	if err != nil {
		message := fmt.Sprintf("Error changing user %v password", e.EmailAddress)
		statusCode := 400
		return message, statusCode, headers
	}

	headers["Authorization"] = fmt.Sprintf("Bearer %v", AccessToken)
	message := fmt.Sprintf("User %v changed password successfully", e.EmailAddress)
	statusCode := 200
	return message, statusCode, headers
}
