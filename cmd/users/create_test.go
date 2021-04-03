package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cognitoidentityproviderif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockAdminCreateUser struct {
	cognitoidentityproviderif.CognitoIdentityProviderAPI
	Response *cognitoidentityprovider.AdminCreateUserOutput
	Error    error
}

func (m mockAdminCreateUser) AdminCreateUser(*cognitoidentityprovider.AdminCreateUserInput) (*cognitoidentityprovider.AdminCreateUserOutput, error) {
	return m.Response, nil
}

func TestCreateUser(t *testing.T) {
	t.Run("Successfully create user", func(t *testing.T) {
		idpMock := mockAdminCreateUser{
			Response: &cognitoidentityprovider.AdminCreateUserOutput{
				User: &cognitoidentityprovider.UserType{
					Username: aws.String("12345"),
				},
			},
			Error: nil,
		}

		app := application{config: configuration{
			UserPoolID: "test",
			IDP:        idpMock,
		}}
		tenantID := "12345"

		_, err := app.createUser(createUserEvent{EmailAddress: "user@example.com"}, tenantID)
		if err != nil {
			t.Fatal("User should have been created")
		}
	})
}
