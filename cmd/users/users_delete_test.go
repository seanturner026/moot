package main

import (
	"testing"

	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminDeleteUser struct {
	cidpif.CognitoIdentityProviderAPI
	Response *cidp.AdminDeleteUserOutput
	Error    error
}

func (m mockAdminDeleteUser) AdminDeleteUser(*cidp.AdminDeleteUserInput) (*cidp.AdminDeleteUserOutput, error) {
	return m.Response, nil
}

func TestDeleteUser(t *testing.T) {
	t.Run("Successfully delete user", func(t *testing.T) {
		idpMock := mockAdminDeleteUser{
			Response: &cidp.AdminDeleteUserOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			UserPoolID: "test",
			idp:        idpMock,
		}}

		err := app.deleteUser(deleteUserEvent{EmailAddress: "user@example.com"})
		if err != nil {
			t.Fatal("User should have been deleted")
		}
	})
}
