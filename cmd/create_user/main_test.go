package main

import (
	"testing"

	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminCreateUser struct {
	cidpif.CognitoIdentityProviderAPI
	Response *cidp.AdminCreateUserOutput
	Error    error
}

func (m mockAdminCreateUser) AdminCreateUser(*cidp.AdminCreateUserInput) (*cidp.AdminCreateUserOutput, error) {
	return m.Response, nil
}

func TestCreateUser(t *testing.T) {
	t.Run("Successfully create user", func(t *testing.T) {
		idpMock := mockAdminCreateUser{
			Response: &cidp.AdminCreateUserOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			UserPoolID: "test",
			idp:        idpMock,
		}}

		err := app.createUser(createUserEvent{EmailAddress: "user@example.com"})
		if err != nil {
			t.Fatal("User should have been created")
		}
	})
}
