package main

// import (
// 	"testing"

// 	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
// 	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
// )

// type mockDescribeUserPoolClient struct {
// 	cidpif.CognitoIdentityProviderAPI
// 	Response *cidp.DescribeUserPoolClientOutput
// 	Error    error
// }

// // type mockInitiateAuth struct {
// // 	cidpif.CognitoIdentityProviderAPI
// // 	Response *cidp.InitiateAuthOutput
// // 	Error    error
// // }

// func (m mockDescribeUserPoolClient) DescribeUserPoolClient(*cidp.DescribeUserPoolClientInput) (*cidp.DescribeUserPoolClientOutput, error) {
// 	return m.Response, nil
// }

// // func (m mockInitiateAuth) InitiateAuth(*cidp.InitiateAuthInput) (*cidp.InitiateAuthOutput, error) {
// // 	return m.Response, nil
// // }

// func TestGetUserPoolClientSecret(t *testing.T) {
// 	t.Run("Successfully obtained client pool secret", func(t *testing.T) {
// 		idpMock := mockDescribeUserPoolClient{
// 			Response: &cidp.DescribeUserPoolClientOutput{},
// 			Error:    nil,
// 		}

// 		app := application{config: configuration{
// 			ClientPoolID: "test",
// 			UserPoolID:   "test",
// 			idp:          idpMock,
// 		}}

// 		_, err := app.getUserPoolClientSecret()
// 		if err != nil {
// 			t.Fatal("App secret should have been obtained")
// 		}
// 	})
// }

// // func TestLoginUser(t *testing.T) {
// // 	t.Run("Successfully logged in user", func(t *testing.T) {
// // 		idpMock := mockInitiateAuth{
// // 			Response: &cidp.InitiateAuthOutput{},
// // 			Error:    nil,
// // 		}

// // 		app := application{config: configuration{
// // 			ClientPoolID: "test",
// // 			UserPoolID:   "test",
// // 			idp:          idpMock,
// // 		}}

// // 		event := loginUserEvent{
// // 			EmailAddress: "user@example.com",
// // 			Password:     "example123$%^",
// // 		}

// // 		_, err := app.loginUser(event, "secretHashExample")
// // 		if err != nil {
// // 			t.Fatal("User should have been logged in")
// // 		}
// // 	})
// // }
