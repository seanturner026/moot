package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/urfave/cli/v2"
)

func getAdminUserID(adminUserEmail, userPoolID string) string {
	input := &cognitoidentityprovider.ListUsersInput{
		AttributesToGet: []*string{aws.String("email")},
		Filter:          aws.String(fmt.Sprintf("email = \"%s\"", adminUserEmail)),
		UserPoolId:      aws.String(userPoolID),
	}

	client := cognitoidentityprovider.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("us-east-1")})))
	resp, err := client.ListUsers(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			fmt.Printf("%v\n", aerr.Error())
		} else {
			fmt.Printf("%v\n", err.Error())
		}
	}
	return *resp.Users[0].Username
}

func printUserID(adminUserID string) {

	type externalDataResponse struct {
		UserID string `json:"user_id"`
	}

	response := &externalDataResponse{}
	response.UserID = adminUserID
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
	}
	_, err = os.Stdout.Write(responseJSON)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "a",
				Aliases:  []string{"admin-user-email"},
				Usage:    "Email address of the Dashboard Admin.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "u",
				Aliases:  []string{"user-pool-id"},
				Usage:    "ID of the User Pool.",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			adminUserEmail := c.String("a")
			userPoolID := c.String("u")
			adminUserID := getAdminUserID(adminUserEmail, userPoolID)
			printUserID(adminUserID)
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
