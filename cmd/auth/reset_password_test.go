package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminRespondToAuthChallenge struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.AdminRespondToAuthChallengeOutput
	Error    error
}

func (m mockAdminRespondToAuthChallenge) AdminRespondToAuthChallenge(*cognitoidentityprovider.AdminRespondToAuthChallengeInput) (*cognitoidentityprovider.AdminRespondToAuthChallengeOutput, error) {
	return m.Response, nil
}

func TestResetPassword(t *testing.T) {
	t.Run("Successfully reset user password", func(t *testing.T) {

		idpMock := mockAdminRespondToAuthChallenge{
			Response: &cognitoidentityprovider.AdminRespondToAuthChallengeOutput{
				AuthenticationResult: &cognitoidentityprovider.AuthenticationResultType{
					AccessToken: aws.String("example"),
				},
			},
			Error: nil,
		}

		app := application{
			Config: configuration{
				ClientPoolID:     "test",
				UserPoolID:       "test",
				ClientPoolSecret: "test",
			},
			IDP: idpMock,
		}

		event := resetPasswordEvent{
			EmailAddress: "user@example.com",
			NewPassword:  "example123$%^",
			SessionID:    "example",
		}

		_, err := app.resetPassword(event, "exampleSecretHash")
		if err != nil {
			t.Fatal("User password should have been reset")
		}
	})
}
