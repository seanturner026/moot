package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminRespondToAuthChallenge struct {
	cidpif.CognitoIdentityProviderAPI
	Response *cidp.AdminRespondToAuthChallengeOutput
	Error    error
}

func (m mockAdminRespondToAuthChallenge) AdminRespondToAuthChallenge(*cidp.AdminRespondToAuthChallengeInput) (*cidp.AdminRespondToAuthChallengeOutput, error) {
	return m.Response, nil
}

func TestResetPassword(t *testing.T) {
	t.Run("Successfully reset user password", func(t *testing.T) {

		idpMock := mockAdminRespondToAuthChallenge{
			Response: &cidp.AdminRespondToAuthChallengeOutput{
				AuthenticationResult: &cidp.AuthenticationResultType{
					AccessToken: aws.String("example"),
				},
			},
			Error: nil,
		}

		app := application{config: configuration{
			ClientPoolID:     "test",
			UserPoolID:       "test",
			ClientPoolSecret: "test",
			idp:              idpMock,
		}}

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
