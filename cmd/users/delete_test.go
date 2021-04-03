package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminDeleteUser struct {
	cognitoidentityprovideriface.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.AdminDeleteUserOutput
	Error    error
}

func (m mockAdminDeleteUser) AdminDeleteUser(*cognitoidentityprovider.AdminDeleteUserInput) (*cognitoidentityprovider.AdminDeleteUserOutput, error) {
	return m.Response, nil
}

func TestDeleteUserFromCognito(t *testing.T) {
	t.Run("Successfully delete user", func(t *testing.T) {
		idpMock := mockAdminDeleteUser{
			Response: &cognitoidentityprovider.AdminDeleteUserOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			UserPoolID: "test",
			IDP:        idpMock,
		}}

		err := app.deleteUserFromCognito(deleteUserEvent{EmailAddress: "user@example.com"})
		if err != nil {
			t.Fatal("User should have been deleted")
		}
	})
}
