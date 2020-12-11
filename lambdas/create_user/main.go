package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/uuid"
)

type createUserEvent struct {
	EmailAddress string `json:"email_address"`
}

// response is of type APIGatewayProxyResponse which leverages the AWS Lambda Proxy Request
// functionality (default behavior)
type response events.APIGatewayProxyResponse

var client *cognitoidentityprovider.CognitoIdentityProvider

func init() {
	client = cognitoidentityprovider.New(session.New())
}

func generatePassword() string {
	id := uuid.New().String()

	return id
}

func createAdminUser(c createUserEvent, id string) {
	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:             aws.String(os.Getenv("USER_POOL_ID")),
		Username:               aws.String(c.EmailAddress),
		DesiredDeliveryMediums: []*string{aws.String("EMAIL")},
		TemporaryPassword:      aws.String(id),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(c.EmailAddress),
			},
		},
	}
	_, err := client.AdminCreateUser(input)
	if err != nil {
		log.Fatalf("[ERROR] Unable to create new admin user, %v", err)
	}
}

func handler(ctx context.Context, c createUserEvent) (response, error) {
	id := generatePassword()
	createAdminUser(c, id)
	body, err := json.Marshal(map[string]interface{}{
		"message": fmt.Sprintf("Created new admin user %v", c.EmailAddress),
	})

	if err != nil {
		return response{StatusCode: 404}, err
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	resp := response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
