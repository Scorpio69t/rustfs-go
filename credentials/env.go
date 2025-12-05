package credentials

import (
	"os"
)

// EnvAWS retrieves credentials from the environment variables
// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_SESSION_TOKEN.
type EnvAWS struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// Retrieve retrieves the credentials from the environment.
func (e *EnvAWS) Retrieve() (Value, error) {
	e.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	e.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	e.SessionToken = os.Getenv("AWS_SESSION_TOKEN")

	return Value{
		AccessKeyID:     e.AccessKeyID,
		SecretAccessKey: e.SecretAccessKey,
		SessionToken:    e.SessionToken,
		SignerType:      SignatureV4,
	}, nil
}

// IsExpired returns if the credentials are expired.
func (e *EnvAWS) IsExpired() bool {
	return false
}

// NewEnvAWS returns a pointer to a new Credentials object
// wrapping the environment variable provider.
func NewEnvAWS() *Credentials {
	return New(&EnvAWS{})
}
