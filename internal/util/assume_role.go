package util

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AssumeTenantRole takes a roleARN and returns session credentials that can be used for
// downstream aws API calls
func AssumeTenantRole(roleARN string) (*session.Session, *credentials.Credentials) {
	session := session.Must(session.NewSession())
	credentials := stscreds.NewCredentials(session, roleARN)
	return session, credentials
}
