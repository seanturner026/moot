package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	cidpif "github.com/aws/aws-sdk-go/service/cognitoidentityprovider/cognitoidentityprovideriface"
	util "github.com/seanturner026/serverless-release-dashboard/pkg/util"
)

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

type application struct {
	config configuration
}

type configuration struct {
	UserPoolID string
	idp        cidpif.CognitoIdentityProviderAPI
}

func (app *application) createUser(e createUserEvent) error {
	input := &cidp.AdminCreateUserInput{
		UserPoolId:             aws.String(app.config.UserPoolID),
		Username:               aws.String(e.EmailAddress),
		DesiredDeliveryMediums: aws.StringSlice([]string{"EMAIL"}),
		ForceAliasCreation:     aws.Bool(true),
		UserAttributes: []*cidp.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(e.EmailAddress),
			},
		},
	}
	_, err := app.config.idp.AdminCreateUser(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			log.Printf("[ERROR] %v", aerr.Error())
		} else {
			log.Printf("[ERROR] %v", err.Error())
		}
		return err
	}
	log.Printf("[INFO] Created new user %v successfully", e.EmailAddress)
	return nil
}

func (app *application) handler(e createUserEvent) (events.APIGatewayProxyResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	err := app.createUser(e)
	if err != nil {
		resp := util.GenerateResponseBody(fmt.Sprintf("Error creating user %v", e.EmailAddress), 404, err, headers)
		return resp, nil
	}

	resp := util.GenerateResponseBody(fmt.Sprintf("Created new user %v", e.EmailAddress), 200, nil, headers)
	return resp, nil
}

func main() {
	config := configuration{
		UserPoolID: os.Getenv("USER_POOL_ID"),
		idp:        cidp.New(session.Must(session.NewSession())),
	}

	app := application{config: config}

	lambda.Start(app.handler)
}
