package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/seanturner026/serverless-release-dashboard/internal/util"
)

type userAuthEvent struct {
	EmailAddress string `dynamodbav:"EmailAddress" json:"email_address"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TenantName   string `dynamodbav:"PK" json:"tenant_name"`
}

type tenantLookupResponse struct {
	ID string `dynamodbav:"OrganizationID" json:"id"`
}

type userAuthResponse struct {
	AccessToken         string    `json:"access_token,omitempty"`
	IDToken             string    `json:"id_token,omitempty"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	ExpiresAt           time.Time `json:"expires_at,omitempty"`
	NewPasswordRequired bool
	SessionID           string `json:"session_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
}

func (app application) generateAuthInput(e userAuthEvent, path string, secretHash string) *cidp.InitiateAuthInput {
	input := &cidp.InitiateAuthInput{}
	input.ClientId = aws.String(app.config.ClientPoolID)
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

func (app application) loginUser(e userAuthEvent, input *cidp.InitiateAuthInput) (userAuthResponse, error) {
	loginUserResp := userAuthResponse{}
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
	loginUserResp.IDToken = *resp.AuthenticationResult.IdToken
	loginUserResp.RefreshToken = *resp.AuthenticationResult.RefreshToken
	loginUserResp.NewPasswordRequired = false
	return loginUserResp, nil
}

func (app application) getTenantID(organizationName string) (string, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("organization"),
			},
			"SK": {
				S: aws.String(organizationName),
			},
		},
		ConsistentRead:         aws.Bool(false),
		ProjectionExpression:   aws.String("OrganizationID"),
		ReturnConsumedCapacity: aws.String("NONE"),
		TableName:              aws.String(app.config.TableName),
	}
	resp, err := app.config.db.GetItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return "", err
	}

	organizationID := &tenantLookupResponse{}
	err = dynamodbattribute.UnmarshalMap(resp.Item, organizationID)
	if err != nil {
		log.Printf("[ERROR] unable to unmarshal dyanmodb response into organizationID")
	}
	return organizationID.ID, nil
}

func (app application) authLoginHandler(event events.APIGatewayV2HTTPRequest, headers map[string]string) (string, int, map[string]string) {
	e := userAuthEvent{}
	err := json.Unmarshal([]byte(event.Body), &e)
	if err != nil {
		log.Printf("[ERROR] %v", err)
	}

	secretHash := util.GenerateSecretHash(app.config.ClientPoolSecret, e.EmailAddress, app.config.ClientPoolID)
	input := app.generateAuthInput(e, event.RawPath, secretHash)
	loginUserResp, err := app.loginUser(e, input)
	if err != nil {
		message := fmt.Sprintf("Error authenticating user %v", e.EmailAddress)
		statusCode := 400
		return message, statusCode, headers
	} else if loginUserResp.NewPasswordRequired {
		headers["X-Session-Id"] = loginUserResp.SessionID
		message := fmt.Sprintf("User %v authorized successfully, password change required", e.EmailAddress)
		statusCode := 200
		return message, statusCode, headers
	}

	tenantID, err := app.getTenantID(e.TenantName)
	if err != nil {
		return fmt.Sprintf("Organization Name %v is invalid", e.TenantName), 400, headers
	}
	tokenTenantID := util.ExtractTenantID(loginUserResp.IDToken)
	if err != nil {
		return fmt.Sprintf("Organization Name %v is invalid", e.TenantName), 400, headers
	}

	if tenantID != tokenTenantID {
		message := "User not authenticated, organisation name is invalid."
		statusCode := 400
		return message, statusCode, headers
	}

	// cookies := []string{
	// 	fmt.Sprintf("Bearer %v; Secure; HttpOnly; SameSite=Strict; Expires=%v", loginUserResp.AccessToken, loginUserResp.ExpiresAt),
	// 	fmt.Sprintf("X-Refresh-Token %v", loginUserResp.RefreshToken),
	// }

	headers["Authorization"] = fmt.Sprintf("Bearer %v", loginUserResp.AccessToken)
	headers["X-Identity-Token"] = loginUserResp.IDToken
	headers["X-Refresh-Token"] = loginUserResp.RefreshToken
	message := fmt.Sprintf("User %v authorized successfully", e.EmailAddress)
	statusCode := 200
	return message, statusCode, headers
}
