package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
)

type mockListUsers struct {
	cidpif.CognitoIdentityProviderAPI
	Response *cidp.ListUsersOutput
	Error    error
}

func (m mockListUsers) ListUsers(*cidp.ListUsersInput) (*cidp.ListUsersOutput, error) {
	return m.Response, nil
}

func TestListUsers(t *testing.T) {
	t.Run("Successfully listed users", func(t *testing.T) {

		idpMock := mockListUsers{
			Response: &cidp.ListUsersOutput{},
			Error:    nil,
		}

		app := application{config: configuration{
			UserPoolID: "test",
			idp:        idpMock,
		}}

		_, err := app.listUsers()
		if err != nil {
			t.Fatal("Users should have been listed")
		}
	})
}

func TestGenerateListUsersResponse(t *testing.T) {
	t.Run("Successfully generated listUsersResponse", func(t *testing.T) {

		idpMock := mockListUsers{
			Response: &cidp.ListUsersOutput{
				Users: []*cidp.UserType{
					{
						Attributes: []*cidp.AttributeType{{
							Value: aws.String("example1"),
						}},
						Username: aws.String("1234-1234-1234-1234"),
					},
					{
						Attributes: []*cidp.AttributeType{{
							Value: aws.String("example2"),
						}},
						Username: aws.String("1234-1234-1234-1234"),
					},
				},
			},
			Error: nil,
		}

		listUsersResponse := generateListUsersResponse(idpMock.Response.Users)
		if listUsersResponse.Users[0].Email != "example1" || listUsersResponse.Users[1].Email != "example2" {
			t.Fatal("Usernames should have been written to listUsersResponse")
		}
	})
}

func TestAppendUserToResponse(t *testing.T) {
	t.Run("Successfully appended user to listUsersResponse", func(t *testing.T) {

		idpMock := mockListUsers{
			Response: &cidp.ListUsersOutput{
				Users: []*cidp.UserType{{
					Attributes: []*cidp.AttributeType{{
						Value: aws.String("example"),
					}},
					Username: aws.String("1234-1234-1234-1234"),
				}},
			},
			Error: nil,
		}

		listUsersResponseMock := &listUsersResponse{}
		userNameMock := user{
			Email: *idpMock.Response.Users[0].Attributes[0].Value,
			ID:    *idpMock.Response.Users[0].Username,
		}
		listUsersResponseMock.appendUserToResponse(userNameMock)
		if listUsersResponseMock.Users[0].Email != "example" {
			t.Fatal("Username should have been appended")
		}
	})
}
