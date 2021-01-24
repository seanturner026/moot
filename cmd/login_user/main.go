package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type loginUserEvent struct {
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
}

type application struct {
	config configuration
}

type configuration struct {
	ClientPoolID     string
	UserPoolID       string
	ClientPoolSecret string
	idp              cidpif.CognitoIdentityProviderAPI
}

type loginUserResponse struct {
	AccessToken         string    `json:"access_token,omitempty"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	ExpiresAt           time.Time `json:"expires_at,omitempty"`
	NewPasswordRequired bool
	SessionID           string `json:"session_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
}

func (app application) loginUser(e loginUserEvent, secretHash string) (loginUserResponse, error) {
	input := &cidp.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME":    aws.String(e.EmailAddress),
			"PASSWORD":    aws.String(e.Password),
			"SECRET_HASH": aws.String(secretHash),
		},
		ClientId: aws.String(app.config.ClientPoolID),
	}

	loginUserResp := loginUserResponse{}
	resp, err := app.config.idp.InitiateAuth(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		loginUserResp.NewPasswordRequired = false
		return loginUserResp, err
	}

	if aws.StringValue(resp.ChallengeName) == "NEW_PASSWORD_REQUIRED" {
		log.Printf("[INFO] New password required for %v", e.EmailAddress)
		loginUserResp.NewPasswordRequired = true
		loginUserResp.SessionID = *resp.Session
		loginUserResp.UserID = *resp.ChallengeParameters["USER_ID_FOR_SRP"]
		return loginUserResp, nil
	}

	log.Printf("[INFO] Authenticated user %v successfully", e.EmailAddress)

	now := time.Now()
	loginUserResp.ExpiresAt = now.Add(time.Second * time.Duration(*resp.AuthenticationResult.ExpiresIn))
	loginUserResp.AccessToken = *resp.AuthenticationResult.AccessToken
	loginUserResp.RefreshToken = *resp.AuthenticationResult.RefreshToken
	loginUserResp.NewPasswordRequired = false
	return loginUserResp, nil
}

func (app application) handler(event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	e := loginUserEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	secretHash := util.GenerateSecretHash(app.config.ClientPoolSecret, e.EmailAddress, app.config.ClientPoolID)
	loginUserResp, err := app.loginUser(e, secretHash)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error logging user %v in", e.EmailAddress), 404, err, headers, []string{})
		return resp, nil

	} else if loginUserResp.NewPasswordRequired {
		headers["X-Session-Id"] = loginUserResp.SessionID
		resp := util.GenerateResponseBody(
			fmt.Sprintf("User %v logged in successfully, password change required", e.EmailAddress), 200, err, headers, []string{},
		)
		return resp, nil
	}

	// cookies := []string{
	// 	fmt.Sprintf("Bearer %v; Secure; HttpOnly; SameSite=Strict; Expires=%v", loginUserResp.AccessToken, loginUserResp.ExpiresAt),
	// 	fmt.Sprintf("X-Refresh-Token %v", loginUserResp.RefreshToken),
	// }
	headers["Authorization"] = fmt.Sprintf("Bearer %v", loginUserResp.AccessToken)
	headers["X-Refresh-Token"] = loginUserResp.RefreshToken
	resp := util.GenerateResponseBody(fmt.Sprintf("User %v logged in successfully", e.EmailAddress), 200, err, headers, []string{})
	return resp, nil
}

func main() {
	config := configuration{
		ClientPoolID:     os.Getenv("CLIENT_POOL_ID"),
		UserPoolID:       os.Getenv("USER_POOL_ID"),
		ClientPoolSecret: os.Getenv("CLIENT_POOL_SECRET"),
		idp:              cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
