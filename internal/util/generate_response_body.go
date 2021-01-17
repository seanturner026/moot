package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

type responseBody struct {
	Message string `json:"message"`
}

// GenerateResponseBody creates the response sent back to the client depending on the error message and error type
func GenerateResponseBody(message string, statusCode int, err error, headers map[string]string) events.APIGatewayV2HTTPResponse {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			message = fmt.Sprintf("%v, %v", message, aerr.Error())
		} else {
			message = fmt.Sprintf("%v, %v", message, err)
		}
	}

	body, marshalErr := json.Marshal(responseBody{Message: message})
	if marshalErr != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", marshalErr)
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	resp := events.APIGatewayV2HTTPResponse{
		StatusCode:      statusCode,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: false,
	}

	log.Printf("[DEBUG] resp %+v", resp)

	return resp
}
