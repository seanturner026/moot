package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
	log "github.com/sirupsen/logrus"
)

type userAuthEvent struct {
	EmailAddress string `json:"email_address"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type userAuthResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	SessionID    string    `json:"session_id,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
}

func (app application) generateAuthInput(e userAuthEvent, path string, secretHash string) *cognitoidentityprovider.InitiateAuthInput {
	input := &cognitoidentityprovider.InitiateAuthInput{}
	input.ClientId = aws.String(app.Config.ClientPoolID)
	if path == "/auth/login" {
		input.AuthFlow = aws.String("USER_PASSWORD_AUTH")
		input.AuthParameters = map[string]*string{
			"USERNAME":    aws.String(e.EmailAddress),
			"PASSWORD":    aws.String(e.Password),
			"SECRET_HASH": aws.String(secretHash),
		}

	} else {
		input.AuthFlow = aws.String("REFRESH_TOKEN_AUTH")
		input.AuthParameters = map[string]*string{
			"REFRESH_TOKEN": aws.String(e.RefreshToken),
			"SECRET_HASH":   aws.String(secretHash),
		}
	}
	return input
}

func (app application) loginUser(e userAuthEvent, input *cognitoidentityprovider.InitiateAuthInput) (userAuthResponse, bool, error) {
	loginUserResp := userAuthResponse{}
	var newPasswordRequired bool
	resp, err := app.IDP.InitiateAuth(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Error(fmt.Sprintf("%v", aerr.Error()))
		} else {
			log.Error(fmt.Sprintf("%v", err.Error()))
		}
		newPasswordRequired = false
		return loginUserResp, newPasswordRequired, err
	}

	if aws.StringValue(resp.ChallengeName) == "NEW_PASSWORD_REQUIRED" {
		log.Info(fmt.Sprintf("new password required for %v", e.EmailAddress))
		newPasswordRequired = true
		loginUserResp.SessionID = *resp.Session
		loginUserResp.UserID = *resp.ChallengeParameters["USER_ID_FOR_SRP"]
		return loginUserResp, newPasswordRequired, nil
	}
	log.Info(fmt.Sprintf("authenticated user %v successfully", e.EmailAddress))

	now := time.Now()
	loginUserResp.ExpiresAt = now.Add(time.Second * time.Duration(*resp.AuthenticationResult.ExpiresIn))
	loginUserResp.AccessToken = *resp.AuthenticationResult.AccessToken
	loginUserResp.RefreshToken = *resp.AuthenticationResult.RefreshToken
	newPasswordRequired = false
	return loginUserResp, newPasswordRequired, nil
}

func (app application) authLoginHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) (string, int, map[string]string) {
	e := userAuthEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Error(fmt.Sprintf("%v", err))
	}

	secretHash := util.GenerateSecretHash(app.Config.ClientPoolSecret, e.EmailAddress, app.Config.ClientPoolID)
	input := app.generateAuthInput(e, event.RawPath, secretHash)
	loginUserResp, newPasswordRequired, err := app.loginUser(e, input)
	if err != nil {
		message := fmt.Sprintf("Error authenticating user %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode, headers
	} else if newPasswordRequired {
		headers["X-Session-Id"] = loginUserResp.SessionID
		message := fmt.Sprintf("User %v authorized successfully, password change required", e.EmailAddress)
		statusCode := 200
		return message, statusCode, headers
	}

	// cookies := []string{
	// 	fmt.Sprintf("Bearer %v; Secure; HttpOnly; SameSite=Strict; Expires=%v", loginUserResp.AccessToken, loginUserResp.ExpiresAt),
	// 	fmt.Sprintf("X-Refresh-Token %v", loginUserResp.RefreshToken),
	// }

	headers["Authorization"] = fmt.Sprintf("Bearer %v", loginUserResp.AccessToken)
	headers["X-Refresh-Token"] = loginUserResp.RefreshToken
	message := fmt.Sprintf("User %v authorized successfully", e.EmailAddress)
	statusCode := 200
	return message, statusCode, headers
}
