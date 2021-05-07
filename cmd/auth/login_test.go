package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockInitiateAuth struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.InitiateAuthOutput
	Error    error
}

func (m mockInitiateAuth) InitiateAuth(*cognitoidentityprovider.InitiateAuthInput) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	return m.Response, nil
}

func TestLoginUser(t *testing.T) {
	t.Run("Successfully logged in user, user must change password", func(t *testing.T) {

		idpMock := mockInitiateAuth{
			Response: &cognitoidentityprovider.InitiateAuthOutput{
				ChallengeName:       aws.String("NEW_PASSWORD_REQUIRED"),
				Session:             aws.String("test"),
				ChallengeParameters: map[string]*string{"USER_ID_FOR_SRP": aws.String("test")},
			},
			Error: nil,
		}

		app := application{
			Config: configuration{
				ClientPoolID: "test",
				UserPoolID:   "test",
			},
			IDP: idpMock,
		}

		event := userAuthEvent{
			EmailAddress: "user@example.com",
			Password:     "example123$%^",
		}

		input := app.generateAuthInput(event, "/login/user", "exampleSecretHash")
		_, _, err := app.loginUser(event, input)
		if err != nil {
			t.Fatal("User should have been logged in")
		}
	})

	t.Run("Successfully logged in user", func(t *testing.T) {

		idpMock := mockInitiateAuth{
			Response: &cognitoidentityprovider.InitiateAuthOutput{
				ChallengeName: nil,
				AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
					AccessToken:  aws.String("test"),
					RefreshToken: aws.String("test"),
					ExpiresIn:    aws.Int64(1),
				},
			},
			Error: nil,
		}

		app := application{
			Config: configuration{
				ClientPoolID: "test",
				UserPoolID:   "test",
			},
			IDP: idpMock,
		}

		event := userAuthEvent{
			EmailAddress: "user@example.com",
			Password:     "example123$%^",
		}

		input := app.generateAuthInput(event, "/login/user", "example_secret_hash")
		_, _, err := app.loginUser(event, input)
		if err != nil {
			t.Fatal("User should have been logged in")
		}
	})

	t.Run("Successfully refreshed user token", func(t *testing.T) {

		idpMock := mockInitiateAuth{
			Response: &cognitoidentityprovider.InitiateAuthOutput{
				ChallengeName: nil,
				AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
					AccessToken:  aws.String("test"),
					RefreshToken: aws.String("test"),
					ExpiresIn:    aws.Int64(1),
				},
			},
			Error: nil,
		}

		app := application{
			Config: configuration{
				ClientPoolID: "test",
				UserPoolID:   "test",
			},
			IDP: idpMock,
		}

		event := userAuthEvent{
			EmailAddress: "user@example.com",
			Password:     "example123$%^",
		}

		input := app.generateAuthInput(event, "/refresh/token", "example_secret_hash")
		_, _, err := app.loginUser(event, input)
		if err != nil {
			t.Fatal("User should have been logged in")
		}
	})
}
