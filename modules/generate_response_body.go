package modules

import (
	"bytes"
	"encoding/json"
	"log"
)

type responseBody struct {
	Message string `json:"message"`
}

// GenerateResponseBody creates the response sent back to the client depending on the error message
func GenerateResponseBody(message string, statusCode int) (bytes.Buffer, int) {
	var body []byte
	var marshalErr error

	body, marshalErr = json.Marshal(responseBody{Message: message})

	if marshalErr != nil {
		log.Printf("[ERROR] Unable to marshal json for response, %v", marshalErr)
		statusCode = 404
	}

	var buf bytes.Buffer
	json.HTMLEscape(&buf, body)

	return buf, statusCode
}
