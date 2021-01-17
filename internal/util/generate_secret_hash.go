package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// GenerateSecretHash performs crytographic calculations to generate the Cognito secret key for the
// current user
func GenerateSecretHash(clientSecret string, emailAddress string, clientPoolID string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(emailAddress + clientPoolID))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
