package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	util "github.com/seanturner026/serverless-release-dashboard/internal/util"
)

func handler() (events.APIGatewayV2HTTPResponse, error) {
	headers := map[string]string{"Content-Type": "application/json"}
	resp := util.GenerateResponseBody("Authorized", 200, nil, headers)
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
