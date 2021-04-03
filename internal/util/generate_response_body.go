package util

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	log "github.com/sirupsen/logrus"
)

// ResponseBody is contains the response sent to the client
type ResponseBody struct {
	Headers map[string]string `json:"headers,omitempty"`
	Message string            `json:"message,omitempty"`
	Cookies []string          `json:"cookies,omitempty"`
}

// GenerateResponseBody creates the response sent back to the client depending on the error message and error type
func GenerateResponseBody(message string, statusCode int, err error, headers map[string]string, cookies []string) events.APIGatewayV2HTTPResponse {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			message = fmt.Sprintf("%v, %v", message, aerr.Error())
		} else {
			message = fmt.Sprintf("%v, %v", message, err)
		}
	}

	body, marshalErr := json.Marshal(ResponseBody{
		Headers: headers,
		Message: message,
	})
	if marshalErr != nil {
		log.Error(fmt.Sprintf("unable to marshal response, %v", marshalErr))
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)
	resp := events.APIGatewayV2HTTPResponse{
		StatusCode:      statusCode,
		Cookies:         cookies,
		Headers:         headers,
		Body:            buf.String(),
		IsBase64Encoded: false,
	}

	return resp
}
