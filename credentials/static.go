package credentials

// Static - static credential provider.
type Static struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	SignerType      SignatureType
}

// Retrieve retrieves static credentials.
func (s Static) Retrieve() (Value, error) {
	return Value{
		AccessKeyID:     s.AccessKeyID,
		SecretAccessKey: s.SecretAccessKey,
		SessionToken:    s.SessionToken,
		SignerType:      s.SignerType,
	}, nil
}

// IsExpired returns if the credentials are expired.
func (s Static) IsExpired() bool {
	return false
}

// NewStaticV4 returns a pointer to a new Credentials object
// wrapping a static credentials provider.
func NewStaticV4(accessKeyID, secretAccessKey, sessionToken string) *Credentials {
	return New(Static{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    sessionToken,
		SignerType:      SignatureV4,
	})
}
