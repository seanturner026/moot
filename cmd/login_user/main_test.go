package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockInitiateAuth struct {
	cidpif.CognitoIdentityProviderAPI
	Response *cidp.InitiateAuthOutput
	Error    error
}

func (m mockInitiateAuth) InitiateAuth(*cidp.InitiateAuthInput) (*cidp.InitiateAuthOutput, error) {
	return m.Response, nil
}

func TestLoginUser(t *testing.T) {
	t.Run("Successfully logged in user, user must change password", func(t *testing.T) {

		idpMock := mockInitiateAuth{
			Response: &cidp.InitiateAuthOutput{
				ChallengeName:       aws.String("NEW_PASSWORD_REQUIRED"),
				Session:             aws.String("test"),
				ChallengeParameters: map[string]*string{"USER_ID_FOR_SRP": aws.String("test")},
			},
			Error: nil,
		}

		app := application{config: configuration{
			ClientPoolID: "test",
			UserPoolID:   "test",
			idp:          idpMock,
		}}

		event := loginUserEvent{
			EmailAddress: "user@example.com",
			Password:     "example123$%^",
		}

		_, err := app.loginUser(event, "secretHashExample")
		if err != nil {
			t.Fatal("User should have been logged in")
		}
	})

	t.Run("Successfully logged in user", func(t *testing.T) {

		idpMock := mockInitiateAuth{
			Response: &cidp.InitiateAuthOutput{
				ChallengeName: nil,
				AuthenticationResult: &cidp.AuthenticationResultType{
					AccessToken:  aws.String("test"),
					RefreshToken: aws.String("test"),
					ExpiresIn:    aws.Int64(1),
				},
			},
			Error: nil,
		}

		app := application{config: configuration{
			ClientPoolID: "test",
			UserPoolID:   "test",
			idp:          idpMock,
		}}

		event := loginUserEvent{
			EmailAddress: "user@example.com",
			Password:     "example123$%^",
		}

		_, err := app.loginUser(event, "secretHashExample")
		if err != nil {
			t.Fatal("User should have been logged in")
		}
	})
}
